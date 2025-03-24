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
				dynamic.Register("openapi", dynamic.AnyVersion, &openapi.Config{})
				w := NewConfigWatcher(&static.Config{Configs: []string{`{"openapi":"3.0","info":{"title":"foo"}}`}})

				ch := make(chan dynamic.ConfigEvent, 1)
				w.AddListener(func(config dynamic.ConfigEvent) {
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
				var ch chan dynamic.ConfigEvent
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				w.AddListener(func(e dynamic.ConfigEvent) {
					require.NotNil(t, e.Config.Data)
				})

				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)

				time.Sleep(500 * time.Millisecond)
				ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{10}}, Raw: []byte("foobar")}}
				time.Sleep(5 * time.Millisecond)
				require.Equal(t, "foobar", c.Data)
			},
		},
		{
			name: "read after file changed",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan dynamic.ConfigEvent
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath}, Raw: []byte("foobar")}}
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

				c, err := w.Read(configPath, &parseError{})
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
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: &slow{}}, nil
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
				var chWatcher chan dynamic.ConfigEvent
				w.providers["foo"] = &testprovider{
					start: func(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
						chWatcher = ch
						return nil
					},
				}

				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				ch := make(chan dynamic.ConfigEvent, 1)
				w.AddListener(func(c dynamic.ConfigEvent) {
					ch <- c
				})
				chWatcher <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath}}}

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
		{
			name: "file delete event",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan dynamic.ConfigEvent
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
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

				var deleted *dynamic.Config
				w.AddListener(func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						deleted = e.Config
					}
				})

				time.Sleep(500 * time.Millisecond)
				ch <- dynamic.ConfigEvent{Name: configPath.String(), Event: dynamic.Delete}
				time.Sleep(5 * time.Millisecond)
				require.NotNil(t, deleted)
				require.Equal(t, c, deleted)

			},
		},
		{
			name: "parent file deleted and child is updated",
			test: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPathParent := mustParse("foo://parent.yml")
				configPathChild := mustParse("foo://child.yml")
				var ch chan dynamic.ConfigEvent
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				parent, err := w.Read(configPathParent, nil)
				require.NoError(t, err)
				require.NotNil(t, parent)

				child, err := w.Read(configPathChild, nil)
				require.NoError(t, err)
				require.NotNil(t, parent)
				dynamic.AddRef(parent, child)

				var deleted *dynamic.Config
				w.AddListener(func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						deleted = e.Config
					}
				})

				time.Sleep(500 * time.Millisecond)
				ch <- dynamic.ConfigEvent{Name: configPathParent.String(), Event: dynamic.Delete}
				time.Sleep(5 * time.Millisecond)
				require.NotNil(t, deleted)
				require.Equal(t, parent, deleted)

				update := &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPathChild}}
				update.Info.Checksum = []byte{2}
				ch <- dynamic.ConfigEvent{Name: configPathChild.String(), Config: update, Event: dynamic.Update}
				time.Sleep(5 * time.Millisecond)
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
				w.providers["foo"] = &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
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
				dynamic.Register("openapi", dynamic.AnyVersion, &openapi.Config{})

				w := NewConfigWatcher(&static.Config{})
				var listenerReceived []*dynamic.Config
				w.AddListener(func(e dynamic.ConfigEvent) {
					listenerReceived = append(listenerReceived, e.Config)
				})
				var ch chan dynamic.ConfigEvent
				w.providers["foo"] = &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
					ch = configs
					return nil
				}}
				pool := safe.NewPool(context.Background())
				err := w.Start(pool)
				require.NoError(t, err)

				ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo.json")}, Raw: []byte(`{"openapi": "3.0","info":{"title":"Foo"}}`)}}
				time.Sleep(time.Duration(100) * time.Millisecond)

				require.Len(t, listenerReceived, 1)
				pool.Stop()

				func() {
					defer func() {
						err := recover()
						require.Equal(t, err.(error).Error(), "send on closed channel")
					}()
					ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo.yml")}}}
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
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						return nil
					},
				}
				var ch chan dynamic.ConfigEvent
				w.providers["bar"] = &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						t.Fatal("read should not be called")
						return nil, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
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

				ch <- dynamic.ConfigEvent{Config: wrapped}

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
	User        string
	calledParse bool
}

func (d *data) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	d.calledParse = true
	return nil
}

type parseError struct{}

func (d *parseError) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	return fmt.Errorf("TEST ERROR")
}

type slow struct{}

func (d *slow) Parse(config *dynamic.Config, _ dynamic.Reader) error {
	config.Data = "foo"
	time.Sleep(5 * time.Second)
	return nil
}

type testprovider struct {
	read  func(u *url.URL) (*dynamic.Config, error)
	start func(chan dynamic.ConfigEvent, *safe.Pool) error
}

func (p *testprovider) Read(u *url.URL) (*dynamic.Config, error) {
	if p.read != nil {
		return p.read(u)
	}
	return nil, nil
}

func (p *testprovider) Start(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
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
