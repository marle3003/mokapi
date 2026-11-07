package server_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/safe"
	"mokapi/server"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Reproduction (before fix): the watcher derives its lookup key from the config's own URL and
// thereby mutates it (OmitHost), while other goroutines read it via ConfigInfo.Key().
func TestConfigWatcher_ConfigKey(t *testing.T) {
	logrus.SetOutput(io.Discard)

	startWatcher := func(t *testing.T) (chan dynamic.ConfigEvent, func()) {
		w := server.NewConfigWatcher(&static.Config{})
		var ch chan dynamic.ConfigEvent
		w.AddProvider("file", &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		}})
		pool := safe.NewPool(context.Background())
		require.NoError(t, w.Start(pool))
		return ch, pool.Stop
	}

	t.Run("adding a config does not change its URL or key", func(t *testing.T) {
		ch, stop := startWatcher(t)
		defer stop()

		c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("file:///foo/bar.yml")}, Raw: []byte("foo")}
		keyBefore := c.Info.Key()
		var added atomic.Bool
		c.Listeners.Add("test", func(dynamic.ConfigEvent) { added.Store(true) })

		ch <- dynamic.ConfigEvent{Config: c}
		require.Eventually(t, added.Load, 2*time.Second, 5*time.Millisecond)

		require.Equal(t, "file:///foo/bar.yml", c.Info.Url.String())
		require.Equal(t, keyBefore, c.Info.Key())
	})

	t.Run("reading the key while the watcher adds the config", func(t *testing.T) {
		ch, stop := startWatcher(t)
		defer stop()

		c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("file:///foo/bar.yml")}, Raw: []byte("foo")}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = c.Info.Key()
				time.Sleep(time.Millisecond)
			}
		}()
		ch <- dynamic.ConfigEvent{Config: c}
		wg.Wait()
	})
}

// Read must not check c.Data without holding the entry lock while
// configChanged reparses the same config and thereby writes c.Data.
func TestConfigWatcher_ReadWhileConfigIsReparsed(t *testing.T) {
	logrus.SetOutput(io.Discard)

	w := server.NewConfigWatcher(&static.Config{})
	configPath := mustParse("foo://file.yml")
	var ch chan dynamic.ConfigEvent
	w.AddProvider("foo", &testprovider{
		read: func(u *url.URL) (*dynamic.Config, error) {
			return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u, Checksum: []byte{0}}, Raw: []byte("foo")}, nil
		},
		start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		},
	})
	pool := safe.NewPool(context.Background())
	require.NoError(t, w.Start(pool))
	defer pool.Stop()

	_, err := w.Read(configPath, nil)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 200; i++ {
			ch <- dynamic.ConfigEvent{Config: &dynamic.Config{
				Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{byte(i)}},
				Raw:  []byte("bar"),
			}}
		}
	}()
	for i := 0; i < 2000; i++ {
		_, err = w.Read(configPath, nil)
		require.NoError(t, err)
	}
	wg.Wait()
}

// reparseCalls counts reparseCounter.Parse calls. It is a variable of the test package, because
// the config parser resets the struct on every parse, so the struct can't hold the state.
var reparseCalls atomic.Int32

// reparseCounter fails its second Parse call.
type reparseCounter struct{}

func (r *reparseCounter) Parse(*dynamic.Config, dynamic.Reader) error {
	if reparseCalls.Add(1) == 2 {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func TestConfigWatcher_ReadAfterFailedReparse(t *testing.T) {
	logrus.SetOutput(io.Discard)

	w := server.NewConfigWatcher(&static.Config{})
	configPath := mustParse("foo://file.json")
	var ch chan dynamic.ConfigEvent
	w.AddProvider("foo", &testprovider{
		read: func(u *url.URL) (*dynamic.Config, error) {
			return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u, Checksum: []byte{0}}, Raw: []byte("{}")}, nil
		},
		start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		},
	})
	pool := safe.NewPool(context.Background())
	require.NoError(t, w.Start(pool))
	defer pool.Stop()

	reparseCalls.Store(0)
	_, err := w.Read(configPath, &reparseCounter{})
	require.NoError(t, err)
	require.Equal(t, int32(1), reparseCalls.Load())

	// the second Parse call, triggered by the changed file, fails
	ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{1}}, Raw: []byte("{}")}}
	require.Eventually(t, func() bool { return reparseCalls.Load() == 2 }, 2*time.Second, 5*time.Millisecond)

	// a config whose reparse failed is parsed again on the next Read
	require.Eventually(t, func() bool {
		_, err = w.Read(configPath, nil)
		return err == nil && reparseCalls.Load() == 3
	}, 2*time.Second, 5*time.Millisecond)

	// once parsed successfully, Read does not parse again
	_, err = w.Read(configPath, nil)
	require.NoError(t, err)
	require.Equal(t, int32(3), reparseCalls.Load())
}

// A config delivered by a provider is parsed by the watcher. A later Read must return that data
// instead of parsing again into the caller's value.
func TestConfigWatcher_ReadProviderDeliveredConfig(t *testing.T) {
	logrus.SetOutput(io.Discard)
	dynamic.Register("openapi", dynamic.AnyVersion, &openapi.Config{})

	w := server.NewConfigWatcher(&static.Config{})
	configPath := mustParse("foo://file.json")
	var ch chan dynamic.ConfigEvent
	w.AddProvider("foo", &testprovider{
		start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		},
	})
	processed := make(chan any, 1)
	w.AddListener(func(e dynamic.ConfigEvent) {
		select {
		case processed <- e.Config.Data:
		default:
		}
	})
	pool := safe.NewPool(context.Background())
	require.NoError(t, w.Start(pool))
	defer pool.Stop()

	ch <- dynamic.ConfigEvent{Config: &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{1}},
		Raw:  []byte(`{"openapi": "3.0","info":{"title":"Foo"}}`),
	}}
	var parsed any
	select {
	case parsed = <-processed:
	case <-time.After(2 * time.Second):
		require.Fail(t, "config was not processed")
	}
	require.IsType(t, &openapi.Config{}, parsed)

	c, err := w.Read(configPath, &openapi.Config{})
	require.NoError(t, err)
	require.Same(t, parsed, c.Data)
}

type plainFile struct {
	Foo string `json:"foo"`
}

// A file that matches no config type (e.g. schemas.json) is parsed with Data == nil. A later Read
// with a typed target, as done when resolving a $ref, must still parse into that target.
func TestConfigWatcher_ReadProviderDeliveredFileWithoutConfigType(t *testing.T) {
	logrus.SetOutput(io.Discard)

	w := server.NewConfigWatcher(&static.Config{})
	configPath := mustParse("foo://schemas.json")
	var ch chan dynamic.ConfigEvent
	w.AddProvider("foo", &testprovider{
		start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		},
	})
	pool := safe.NewPool(context.Background())
	require.NoError(t, w.Start(pool))
	defer pool.Stop()

	ch <- dynamic.ConfigEvent{Config: &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: configPath, Checksum: []byte{1}},
		Raw:  []byte(`{"foo":"bar"}`),
	}}
	// Data stays nil, so the watcher sends no event to wait for. It handles events one at a time,
	// so once it accepts a second event, the first one is fully processed.
	ch <- dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("foo://other.json"), Checksum: []byte{2}}}}

	c, err := w.Read(configPath, &plainFile{})
	require.NoError(t, err)
	require.Equal(t, &plainFile{Foo: "bar"}, c.Data)
}
