package dynamic

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"testing"
	"time"
)

func TestConfigWatcher_Read(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "no provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				u := mustParse("file.yml")
				c, err := w.Read(u)
				require.EqualError(t, err, "unsupported scheme: file.yml")
				require.Nil(t, c)
			},
		},
		{
			name: "with provider",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*common.Config, error) {
						require.Equal(t, configPath, u)
						return common.NewConfig(common.ConfigInfo{Url: u}), nil
					},
				}

				c, err := w.Read(configPath)
				require.NoError(t, err)
				require.Equal(t, configPath, c.Info.Url)
			},
		},
		{
			name: "read twice",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*common.Config, error) {
						require.Equal(t, configPath, u)
						return common.NewConfig(common.ConfigInfo{Url: u}), nil
					},
				}

				c1, err := w.Read(configPath)
				require.NoError(t, err)
				require.Equal(t, configPath, c1.Info.Url)

				c2, err := w.Read(configPath)
				require.NoError(t, err)
				require.Equal(t, configPath, c2.Info.Url)
				require.True(t, c1 == c2, "should be same reference")
			},
		},
		{
			name: "provider read error",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*common.Config, error) {
						require.Equal(t, configPath, u)
						return nil, fmt.Errorf("TEST ERROR")
					},
				}

				c, err := w.Read(configPath)
				require.EqualError(t, err, "TEST ERROR")
				require.Nil(t, c)
			},
		},
		{
			name: "file changed after read",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan *common.Config
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*common.Config, error) {
						c := common.NewConfig(common.ConfigInfo{Url: u})
						c.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan *common.Config, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				c, err := w.Read(configPath)
				require.NoError(t, err)
				require.NotNil(t, c)

				ch <- &common.Config{Info: common.ConfigInfo{Url: configPath}, Raw: []byte("foobar"), Checksum: []byte{10}}
				time.Sleep(5 * time.Millisecond)
				require.Equal(t, "foobar", c.Data)
			},
		},
		{
			name: "read after file changed",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan *common.Config
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*common.Config, error) {
						return &common.Config{Info: common.ConfigInfo{Url: u}}, nil
					},
					start: func(configs chan *common.Config, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				}
				pool := safe.NewPool(context.Background())
				w.Start(pool)
				defer pool.Stop()

				ch <- &common.Config{Info: common.ConfigInfo{Url: configPath}, Raw: []byte("foobar")}
				time.Sleep(time.Duration(100) * time.Millisecond)

				c, err := w.Read(configPath)
				require.NoError(t, err)
				require.NotNil(t, c)
				require.Equal(t, "foobar", c.Data)
			},
		},
		{
			name: "config parse error",
			f: func(t *testing.T) {
				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.providers["foo"] = &testprovider{
					read: func(u *url.URL) (*common.Config, error) {
						require.Equal(t, configPath, u)
						return &common.Config{Info: common.ConfigInfo{Url: u}, Data: &data{
							parse: func(config *common.Config, reader common.Reader) error {
								return fmt.Errorf("TEST ERROR")
							},
						}}, nil
					},
				}

				c, err := w.Read(configPath)
				require.EqualError(t, err, "parsing file foo://file.yml: TEST ERROR")
				require.Nil(t, c)
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
				w.providers["foo"] = &testprovider{start: func(configs chan *common.Config, pool *safe.Pool) error {
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
				w := NewConfigWatcher(&static.Config{})
				var listenerReceived []*common.Config
				w.AddListener(func(config *common.Config) {
					listenerReceived = append(listenerReceived, config)
				})
				var ch chan *common.Config
				w.providers["foo"] = &testprovider{start: func(configs chan *common.Config, pool *safe.Pool) error {
					ch = configs
					return nil
				}}
				pool := safe.NewPool(context.Background())
				err := w.Start(pool)
				require.NoError(t, err)

				ch <- &common.Config{Info: common.ConfigInfo{Url: mustParse("foo.yml")}}
				time.Sleep(time.Duration(100) * time.Millisecond)

				require.Len(t, listenerReceived, 1)
				pool.Stop()

				func() {
					defer func() {
						err := recover()
						require.Equal(t, err.(error).Error(), "send on closed channel")
					}()
					ch <- &common.Config{Info: common.ConfigInfo{Url: mustParse("foo.yml")}}
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
								Filename: "foo.yml",
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
								Url: "foo",
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
								Url: "git",
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

type data struct {
	parse func(config *common.Config, reader common.Reader) error
}

type testprovider struct {
	read  func(u *url.URL) (*common.Config, error)
	start func(chan *common.Config, *safe.Pool) error
}

func (d *data) Parse(config *common.Config, reader common.Reader) error {
	if d.parse != nil {
		return d.parse(config, reader)
	}
	return nil
}

func (p *testprovider) Read(u *url.URL) (*common.Config, error) {
	if p.read != nil {
		return p.read(u)
	}
	return nil, nil
}

func (p *testprovider) Start(ch chan *common.Config, pool *safe.Pool) error {
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
