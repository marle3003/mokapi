package server_test

import (
	"context"
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/safe"
	"mokapi/server"
	"mokapi/try"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfigWatcher_Read(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no provider",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
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
				w := server.NewConfigWatcher(&static.Config{Configs: []string{`{"openapi":"3.0","info":{"title":"foo"}}`}})

				ch := make(chan dynamic.ConfigEvent, 1)
				w.AddListener(func(config dynamic.ConfigEvent) {
					ch <- config
				})

				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
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
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
				})

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.Equal(t, configPath, c.Info.Url)
			},
		},
		{
			name: "read twice",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
				})

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
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return nil, fmt.Errorf("TEST ERROR")
					},
				})

				c, err := w.Read(configPath, nil)
				require.EqualError(t, err, "TEST ERROR")
				require.Nil(t, c)
			},
		},
		{
			name: "file changed after read",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				processed := make(chan struct{}, 1)
				w.AddListener(func(e dynamic.ConfigEvent) {
					require.NotNil(t, e.Config.Data)
					select {
					case processed <- struct{}{}:
					default:
					}
				})

				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
				defer pool.Stop()

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)

				time.Sleep(500 * time.Millisecond)
				ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{10}}, Raw: []byte("foobar")}}
				// wait until the watcher finished parsing, otherwise Data is read while it is written
				select {
				case <-processed:
				case <-time.After(2 * time.Second):
					require.Fail(t, "config was not processed")
				}
				current, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.Equal(t, "foobar", current.Data)
			},
		},
		{
			name: "read after file changed",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
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
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
					},
				})

				c, err := w.Read(configPath, &parseError{})
				require.EqualError(t, err, "parsing file foo://file.yml: TEST ERROR")
				require.Nil(t, c)
			},
		},
		{
			name: "reading while parsing",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						require.Equal(t, configPath, u)
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: &slow{}}, nil
					},
				})

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
			name: "read explicit while implicit is reading",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
				defer pool.Stop()
				dynamic.Register("test", nil, &slowUnmarshal{})

				ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath}, Raw: []byte("test: 1.0")}}

				time.Sleep(500 * time.Millisecond)

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c.Data)
			},
		},
		{
			name: "should not invoke listeners when content is unknown",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				var chWatcher chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					start: func(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
						chWatcher = ch
						return nil
					},
				})

				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
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
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
				defer pool.Stop()

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)

				var deleted atomic.Pointer[dynamic.Config]
				w.AddListener(func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						deleted.Store(e.Config)
					}
				})

				time.Sleep(500 * time.Millisecond)
				ch <- dynamic.ConfigEvent{Name: configPath.String(), Event: dynamic.Delete}
				require.Eventually(t, func() bool { return deleted.Load() != nil }, 2*time.Second, 5*time.Millisecond)
				require.Equal(t, c, deleted.Load())

			},
		},
		{
			name: "parent file deleted and child is updated",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPathParent := mustParse("foo://parent.yml")
				configPathChild := mustParse("foo://child.yml")
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
				defer pool.Stop()

				parent, err := w.Read(configPathParent, nil)
				require.NoError(t, err)
				require.NotNil(t, parent)

				child, err := w.Read(configPathChild, nil)
				require.NoError(t, err)
				require.NotNil(t, parent)
				dynamic.AddRef(parent, child)

				var deleted atomic.Pointer[dynamic.Config]
				w.AddListener(func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						deleted.Store(e.Config)
					}
				})

				time.Sleep(500 * time.Millisecond)
				ch <- dynamic.ConfigEvent{Name: configPathParent.String(), Event: dynamic.Delete}
				require.Eventually(t, func() bool { return deleted.Load() != nil }, 2*time.Second, 5*time.Millisecond)
				require.Equal(t, parent, deleted.Load())

				update := &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPathChild}}
				update.Info.Checksum = []byte{2}
				ch <- dynamic.ConfigEvent{Name: configPathChild.String(), Config: update, Event: dynamic.Update}
				time.Sleep(5 * time.Millisecond)
			},
		},
		{
			name: "reading a referenced file before reading it as main file",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				configPath := mustParse("foo://file.yml")
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
				defer pool.Stop()

				c, err := w.Read(configPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)

				var mainFile atomic.Pointer[dynamic.Config]
				w.AddListener(func(e dynamic.ConfigEvent) {
					mainFile.Store(e.Config)
				})

				// The listener should be called even checksum matches to the specific w.Read call
				// otherwise the configuration file will be ignored.
				ch <- dynamic.ConfigEvent{Name: configPath.String(), Event: dynamic.Create, Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath, Checksum: c.Info.Checksum}}}
				require.Eventually(t, func() bool { return mainFile.Load() != nil }, 2*time.Second, 5*time.Millisecond)
				require.Equal(t, c, mainFile.Load())
			},
		},
		{
			name: "provider triggers main file and then referenced file",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				mainPath := mustParse("foo://file.yml")
				mainConfig := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mainPath, Checksum: []byte{1}}}
				refPath := mustParse("foo://ref.yml")
				refConfig := &dynamic.Config{Info: dynamic.ConfigInfo{Url: refPath, Checksum: []byte{1}}}

				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
						c.Info.Checksum = []byte{1}
						return c, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
				pool := safe.NewPool(context.Background())
				_ = w.Start(pool)
				defer pool.Stop()

				// listeners are called concurrently
				var m sync.Mutex
				var events []string
				w.AddListener(func(e dynamic.ConfigEvent) {
					m.Lock()
					defer m.Unlock()
					events = append(events, e.Config.Info.Path())
				})
				receivedEvents := func(n int) []string {
					require.Eventually(t, func() bool {
						m.Lock()
						defer m.Unlock()
						return len(events) >= n
					}, time.Second, time.Millisecond)
					m.Lock()
					defer m.Unlock()
					return append([]string{}, events...)
				}

				ch <- dynamic.ConfigEvent{Name: mainPath.String(), Event: dynamic.Create, Config: mainConfig}
				c, err := w.Read(refPath, nil)
				require.NoError(t, err)
				require.NotNil(t, c)
				dynamic.AddRef(mainConfig, c)

				ch <- dynamic.ConfigEvent{Name: refPath.String(), Event: dynamic.Create, Config: refConfig}

				require.ElementsMatch(t, []string{"foo://file.yml", "foo://ref.yml"}, receivedEvents(2))

				refUpdate := &dynamic.Config{Info: dynamic.ConfigInfo{Url: refPath, Checksum: []byte{2}}}
				ch <- dynamic.ConfigEvent{Name: refPath.String(), Event: dynamic.Create, Config: refUpdate}

				require.ElementsMatch(t, []string{"foo://file.yml", "foo://ref.yml", "foo://file.yml", "foo://ref.yml"}, receivedEvents(4))
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
				w := server.NewConfigWatcher(&static.Config{})
				pool := safe.NewPool(context.Background())

				err := w.Start(pool)
				require.NoError(t, err)
				pool.Stop()
			},
		},
		{
			name: "provider error",
			f: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				pool := safe.NewPool(context.Background())
				w.AddProvider("foo", &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
					return fmt.Errorf("TEST ERROR")
				}})

				err := w.Start(pool)
				require.EqualError(t, err, "TEST ERROR")
				pool.Stop()
			},
		},
		{
			name: "closing",
			f: func(t *testing.T) {
				dynamic.Register("openapi", dynamic.AnyVersion, &openapi.Config{})

				w := server.NewConfigWatcher(&static.Config{})
				var listenerReceived atomic.Int32
				w.AddListener(func(e dynamic.ConfigEvent) {
					listenerReceived.Add(1)
				})
				var ch chan dynamic.ConfigEvent
				w.AddProvider("foo", &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
					ch = configs
					return nil
				}})
				pool := safe.NewPool(context.Background())
				err := w.Start(pool)
				require.NoError(t, err)

				ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo.json")}, Raw: []byte(`{"openapi": "3.0","info":{"title":"Foo"}}`)}}
				require.Eventually(t, func() bool { return listenerReceived.Load() > 0 }, 2*time.Second, 5*time.Millisecond)
				require.Equal(t, int32(1), listenerReceived.Load())
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

	// no parallel git provider is not thread-safe
	for _, tc := range testcases {
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
				w := server.NewConfigWatcher(&static.Config{})
				w.AddProvider("foo", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						return &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo://foo.yml")}, Raw: []byte("foo")}, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						return nil
					},
				})
				var ch chan dynamic.ConfigEvent
				w.AddProvider("bar", &testprovider{
					read: func(u *url.URL) (*dynamic.Config, error) {
						t.Fatal("read should not be called")
						return nil, nil
					},
					start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
						ch = configs
						return nil
					},
				})
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

func TestConfigWatcher_PanicInListener(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "panic in registered listener does not stop later events",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				var received atomic.Int32
				w.AddListener(func(e dynamic.ConfigEvent) {
					if e.Config.Info.Url.Path == "first.json" {
						panic("TEST PANIC in listener")
					}
					received.Add(1)
				})
				ch := startWatcherWithChannel(t, w)

				sendEvent(t, ch, newPanicTestConfig("first.json", `{}`))
				sendEvent(t, ch, newPanicTestConfig("second.json", `{}`))

				require.Eventually(t, func() bool { return received.Load() == 1 }, 2*time.Second, 5*time.Millisecond)
			},
		},
		{
			name: "panic while parsing does not block later update of same file",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				var received atomic.Int32
				w.AddListener(func(e dynamic.ConfigEvent) {
					received.Add(1)
				})
				ch := startWatcherWithChannel(t, w)

				sendEvent(t, ch, newPanicTestConfig("foo.json", `{"panic": true}`))
				// the update of the same file locks the entry the panicking parse may have left locked
				sendEvent(t, ch, newPanicTestConfig("foo.json", `{"panic": false}`))

				require.Eventually(t, func() bool { return received.Load() == 1 }, 2*time.Second, 5*time.Millisecond)
			},
		},
		{
			name: "panic in a config listener other than the first does not stop the watcher",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				var received atomic.Int32
				w.AddListener(func(e dynamic.ConfigEvent) { received.Add(1) })
				ch := startWatcherWithChannel(t, w)

				c := newPanicTestConfig("first.json", `{}`)
				panicked := make(chan struct{})
				// the first listener runs synchronously, the following ones in their own goroutine
				c.Listeners.Add("first", func(dynamic.ConfigEvent) {})
				c.Listeners.Add("panic", func(dynamic.ConfigEvent) {
					defer close(panicked)
					panic("TEST PANIC in config listener")
				})
				sendEvent(t, ch, c)
				sendEvent(t, ch, newPanicTestConfig("second.json", `{}`))

				select {
				case <-panicked:
				case <-time.After(2 * time.Second):
					t.Fatal("panicking listener was not called")
				}
				require.Eventually(t, func() bool { return received.Load() == 2 }, 2*time.Second, 5*time.Millisecond)
			},
		},
		{
			name: "panic in a config listener on delete does not skip the other listeners",
			test: func(t *testing.T) {
				w := server.NewConfigWatcher(&static.Config{})
				var removed atomic.Int32
				w.AddListener(func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						removed.Add(1)
					}
				})
				ch := startWatcherWithChannel(t, w)

				c := newPanicTestConfig("first.json", `{}`)
				var after atomic.Int32
				c.Listeners.Add("panic", func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						panic("TEST PANIC in config listener")
					}
				})
				c.Listeners.Add("after", func(e dynamic.ConfigEvent) {
					if e.Event == dynamic.Delete {
						after.Add(1)
					}
				})
				sendEvent(t, ch, c)
				ch <- dynamic.ConfigEvent{Event: dynamic.Delete, Name: "first.json"}

				require.Eventually(t, func() bool { return after.Load() == 1 && removed.Load() == 1 }, 2*time.Second, 5*time.Millisecond)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func startWatcherWithChannel(t *testing.T, w *server.ConfigWatcher) chan dynamic.ConfigEvent {
	var ch chan dynamic.ConfigEvent
	w.AddProvider("foo", &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
		ch = configs
		return nil
	}})
	pool := safe.NewPool(context.Background())
	t.Cleanup(pool.Stop)
	require.NoError(t, w.Start(pool))
	return ch
}

// sendEvent fails instead of hanging when the watcher's event loop is stuck or gone.
func sendEvent(t *testing.T, ch chan dynamic.ConfigEvent, c *dynamic.Config) {
	select {
	case ch <- dynamic.ConfigEvent{Event: dynamic.Create, Config: c}:
	case <-time.After(2 * time.Second):
		t.Fatalf("watcher does not accept event for %v", c.Info.Url)
	}
}

func newPanicTestConfig(name, raw string) *dynamic.Config {
	return &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: mustParse(name), Checksum: []byte(raw)},
		Raw:  []byte(raw),
		Data: &panicOnParse{},
	}
}

// panicOnParse panics in Parse when the config contains "panic": true.
type panicOnParse struct {
	Panic bool `json:"panic"`
}

func (d *panicOnParse) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	if d.Panic {
		panic("TEST PANIC in parse")
	}
	return nil
}

type parseError struct{}

func (d *parseError) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	return fmt.Errorf("TEST ERROR")
}

type slow struct{}

func (d *slow) Parse(config *dynamic.Config, _ dynamic.Reader) error {
	time.Sleep(5 * time.Second)
	config.Data = "foo"
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

type slowUnmarshal struct{}

func (s *slowUnmarshal) UnmarshalYAML(node *yaml.Node) error {
	time.Sleep(5 * time.Second)

	type alias slowUnmarshal
	tmp := alias(*s)
	err := node.Decode(&tmp)
	if err != nil {
		return err
	}
	*s = slowUnmarshal(tmp)
	return nil
}
