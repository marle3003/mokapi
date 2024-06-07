package server

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/safe"
	"mokapi/try"
	"net/url"
	"testing"
	"time"
)

func TestConfigWatcher_Read(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no provider",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				u := mustParse("file.yml")
				c, err := w.Read(u, nil)
				require.EqualError(t, err, "unsupported scheme: file.yml")
				require.Nil(t, c)
			},
		},
		{
			name: "cli configs",
			test: func(t *testing.T) {
				dynamic.Register("openapi", &openapi.Config{})
				w := NewConfigWatcher(&static.Config{Configs: []string{`{"openapi":"3.0","info":{"title":"foo"}}`}})

				ch := make(chan *dynamic.Config, 1)
				w.AddListener(func(config *dynamic.Config) {
					ch <- config
				})

				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				c := try.GetConfig(t, ch, 1*time.Second)
				require.NotNil(t, c)
				require.IsType(t, &openapi.Config{}, c.Data)
				require.Equal(t, "foo", c.Data.(*openapi.Config).Info.Name)
			},
		},
		{
			name: "with provider",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
				}

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.Equal(t, configPath, c.Info.Url)
			},
		},
		{
			name: "read twice",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
				}

				c1, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.Equal(t, configPath, c1.Info.Url)

				c2, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.Equal(t, configPath, c2.Info.Url)
				require.True(t, c1 == c2, "should be same reference")
			},
		},
		{
			name: "provider read error",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return nil, fmt.Errorf("TEST ERROR")
					},
				}

				c, err := w.Read(configPath, nil)
				require.EqualError(t, err, "TEST ERROR")
				require.Nil(t, c)
			},
		},
		{
			name: "file changed after read",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan *dynamic.Config
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)

				time.Sleep(500 * time.Millisecond)
				ch <- &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{10}}, Raw: []byte("foobar")}
				time.Sleep(5 * time.Millisecond)
				require.Equal(t, "foobar", c.Data)
			},
		},
		{
			name: "read after file changed",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan *dynamic.Config
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
					start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				ch <- &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath}, Raw: []byte("foobar")}
				time.Sleep(time.Duration(100) * time.Millisecond)

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)
				require.Equal(t, "foobar", c.Data)
			},
		},
		{
			name: "config parse error",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
				}

				c, err := w.Read(configPath, &data{
					parse: func(config *dynamic.Config, reader dynamic.Reader) error {
						return fmt.Errorf("TEST ERROR")
					}})
				require.EqualError(t, err, "parsing file foo://file.yml: TEST ERROR")
				require.Nil(t, c)
			},
		},
		{
			name: "reading while parsing",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: &data{
							parse: func(config *dynamic.Config, reader dynamic.Reader) error {
								config.Data = "foo"
								time.Sleep(5 * time.Second)
								return nil
							},
						}}, nil
					},
				}

				ch := make(chan interface{}, 2)

				go func() {
					c1, err := w.Read(configPath, nil)
					require.NoError(t, err)
					require.Equal(t, configPath, c1.Info.Url)
					ch <- c1.Data
				}()

				go func() {
					c1, err := w.Read(configPath, nil)
					require.NoError(t, err)
					require.Equal(t, configPath, c1.Info.Url)
					ch <- c1.Data
				}()

				i1 := <-ch
				i2 := <-ch

				require.True(t, i1 == i2, "should be same reference")
			},
		},
		{
			name: "should not invoke listeners when content is unknown",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				var chWatcher chan *dynamic.Config
				w.providers["foo"] = &testprovider{
					start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
						chWatcher = configs
						return nil
					},
				}

				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				ch := make(chan *dynamic.Config, 1)
				w.AddListener(func(config *dynamic.Config) {
					ch <- config
				})
				chWatcher <- &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath}}

			Wait:
				for {
					select {
					case <-ch:
						require.Fail(t, "Unknown config should not trigger listener")
					case <-time.After(1 * time.Second):
						break Wait
					}
				}
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestConfigWatcher_Start(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "no provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				pool := safe.NewPool(context.Background())

				err := w.Start(pool)
				require.NoError(t, err)
				pool.Stop()
			},
		},
		{
			name: "provider error",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				pool := safe.NewPool(context.Background())
				w.providers["foo"] = &testprovider{start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
					return fmt.Errorf("TEST ERROR")
				}}

				err := w.Start(pool)
				require.EqualError(t, err, "TEST ERROR")
				pool.Stop()
			},
		},
		{
			name: "closing",
			f: func(t *testing.T) {
				dynamic.Register("openapi", &openapi.Config{})

				w := NewConfigWatcher(&static.Config{})
				var listenerReceived []*dynamic.Config
				w.AddListener(func(config *dynamic.Config) {
					listenerReceived = append(listenerReceived, config)
				})
				var ch chan *dynamic.Config
				w.providers["foo"] = &testprovider{start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
					ch = configs
					return nil
				}}
				pool := safe.NewPool(context.Background())
				err := w.Start(pool)
				require.NoError(t, err)

				ch <- &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo.json")}, Raw: []byte(`{"openapi": "3.0","info":{"title":"Foo"}}`)}
				time.Sleep(time.Duration(100) * time.Millisecond)

				require.Len(t, listenerReceived, 1)
				pool.Stop()

				func() {
					defer func() {
						err := recover()
						require.Equal(t, err.(error).Error(), "send on closed channel")
					}()
					ch <- &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo.yml")}}
				}()
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}

func TestConfigWatcher_New(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "file provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(
					&static.Config{
						Providers: static.Providers{
							File: static.FileProvider{
								Filenames: []string{"foo.yml"},
							},
						},
					},
				)
				require.Contains(t, w.providers, "file")
			},
		},
		{
			name: "http provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(
					&static.Config{
						Providers: static.Providers{
							Http: static.HttpProvider{
								Urls: []string{"foo"},
							},
						},
					},
				)
				require.Contains(t, w.providers, "http")
			},
		},
		{
			name: "git provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(
					&static.Config{
						Providers: static.Providers{
							Git: static.GitProvider{
								Urls: []string{"git"},
							},
						},
					},
				)
				require.Contains(t, w.providers, "git")
			},
		},
		{
			name: "npm provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(
					&static.Config{
						Providers: static.Providers{
							Npm: static.NpmProvider{
								Packages: []static.NpmPackage{{Name: "foo"}},
							},
						},
					},
				)
				require.Contains(t, w.providers, "npm")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}

func TestConfigWatcher_Wrapping(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "when file is read first by reference and after by (outer) provider",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo://foo.yml")}, Raw: []byte("foo")}, nil
					},
					start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
						return nil
					},
				}
				var ch chan *dynamic.Config
				w.providers["bar"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						t.Fatal("read should not be called")
						return nil, nil
					},
					start: func(configs chan *dynamic.Config, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				c, err := w.Read(mustParse("foo://foo.yml"), nil)
				require.NoError(t, err)
				require.Equal(t, "foo", string(c.Raw))
				require.Equal(t, "foo://foo.yml", c.Info.Url.String())

				wrapped := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo://foo.yml")}, Raw: []byte("foo")}
				dynamic.Wrap(dynamic.ConfigInfo{Url: mustParse("bar://foo.yml")}, wrapped)

				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				err = w.Start(pool)
				require.NoError(t, err)

				ch <- wrapped

				time.Sleep(2 * time.Second)

				c, err = w.Read(mustParse("bar://foo.yml"), nil)
				require.NoError(t, err)
				require.Equal(t, "bar://foo.yml", c.Info.Url.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

type data struct {
	parse func(config *dynamic.Config, reader dynamic.Reader) error
}

type testprovider struct {
	read  func(u *url.URL) (*dynamic.Config, error)
	start func(chan *dynamic.Config, *safe.Pool) error
}

func (d *data) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if d.parse != nil {
		return d.parse(config, reader)
	}
	return nil
}

func (p *testprovider) Read(u *url.URL) (*dynamic.Config, error) {
	if p.read != nil {
		return p.read(u)
	}
	return nil, nil
}

func (p *testprovider) Start(ch chan *dynamic.Config, pool *safe.Pool) error {
	if p.start != nil {
		return p.start(ch, pool)
	}
	return nil
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
