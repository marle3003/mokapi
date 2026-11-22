package server_test

import (
	"context"
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
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

// refChild is a file that is referenced by refParent, like a schema file by an OpenAPI file.
type refChild struct {
	Name string `json:"name"`
}

var refParentParses atomic.Int32

// refParent reads refChild in its Parse hook and registers it as a reference. It counts its
// parses in a variable of the test package, because every parse works on a fresh object.
type refParent struct {
	Child *refChild
}

func (p *refParent) Parse(c *dynamic.Config, r dynamic.Reader) error {
	refParentParses.Add(1)
	ref, err := r.Read(mustParse("foo://child.json"), &refChild{})
	if err != nil {
		return err
	}
	dynamic.AddRef(c, ref)
	p.Child = ref.Data().(*refChild)
	return nil
}

// refFiles starts a watcher whose provider serves parent.json and child.json.
// The returned function changes the content of child.json as the provider would do.
func refFiles(t *testing.T) (w *server.ConfigWatcher, changeChild func(name string, checksum byte)) {
	t.Helper()
	logrus.SetOutput(io.Discard)

	w = server.NewConfigWatcher(&static.Config{})
	var ch chan dynamic.ConfigEvent
	w.AddProvider("foo", &testprovider{
		read: func(u *url.URL) (*dynamic.Config, error) {
			raw := `{}`
			if u.Host == "child.json" {
				raw = `{"name":"v1"}`
			}
			return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u, Checksum: []byte{1}}, Raw: []byte(raw)}, nil
		},
		start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		},
	})
	pool := safe.NewPool(context.Background())
	require.NoError(t, w.Start(pool))
	t.Cleanup(pool.Stop)

	refParentParses.Store(0)
	return w, func(name string, checksum byte) {
		ch <- dynamic.ConfigEvent{Event: dynamic.Update, Config: &dynamic.Config{
			Info: dynamic.ConfigInfo{Url: mustParse("foo://child.json"), Checksum: []byte{checksum}},
			Raw:  []byte(fmt.Sprintf(`{"name":%q}`, name)),
		}}
	}
}

func TestConfigWatcher_ReferencedFileChanged(t *testing.T) {
	t.Run("referencing file is reparsed and refers to the new data", func(t *testing.T) {
		w, changeChild := refFiles(t)
		parent, err := w.Read(mustParse("foo://parent.json"), &refParent{})
		require.NoError(t, err)
		oldChild := parent.Data().(*refParent).Child
		require.Equal(t, "v1", oldChild.Name)

		changeChild("v2", 2)

		require.Eventually(t, func() bool { return refParentParses.Load() == 2 }, 2*time.Second, 5*time.Millisecond)
		require.Eventually(t, func() bool {
			return parent.Data().(*refParent).Child.Name == "v2"
		}, 2*time.Second, 5*time.Millisecond)
		require.Equal(t, "v1", oldChild.Name, "data of the first parse must not change")
	})

	t.Run("reference and listener are registered on the real config, not on a view", func(t *testing.T) {
		w, changeChild := refFiles(t)
		parent, err := w.Read(mustParse("foo://parent.json"), &refParent{})
		require.NoError(t, err)
		child, err := w.Read(mustParse("foo://child.json"), nil)
		require.NoError(t, err)

		require.Equal(t, []*dynamic.Config{child}, parent.Refs.List(false))

		// the listener of the first parse is still active after a reparse
		changeChild("v2", 2)
		require.Eventually(t, func() bool { return refParentParses.Load() == 2 }, 2*time.Second, 5*time.Millisecond)
		changeChild("v3", 3)
		require.Eventually(t, func() bool { return refParentParses.Load() == 3 }, 2*time.Second, 5*time.Millisecond)
		require.Equal(t, []*dynamic.Config{child}, parent.Refs.List(false))
	})

	t.Run("reading the referencing file while the referenced file is reparsed", func(t *testing.T) {
		w, changeChild := refFiles(t)
		parentUrl := mustParse("foo://parent.json")
		_, err := w.Read(parentUrl, &refParent{})
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 2; i <= 100; i++ {
				changeChild(fmt.Sprintf("v%d", i), byte(i))
			}
		}()
		for i := 0; i < 2000; i++ {
			parent, err := w.Read(parentUrl, nil)
			require.NoError(t, err)
			if p, ok := parent.Data().(*refParent); ok {
				require.NotEmpty(t, p.Child.Name)
			}
		}
		wg.Wait()
	})
}

func TestConfigWatcher_ReadTargetIsOnlyATypeHint(t *testing.T) {
	logrus.SetOutput(io.Discard)

	w := server.NewConfigWatcher(&static.Config{})
	w.AddProvider("foo", &testprovider{
		read: func(u *url.URL) (*dynamic.Config, error) {
			return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u, Checksum: []byte{1}}, Raw: []byte(`{"name":"foo"}`)}, nil
		},
	})

	target := &refChild{}
	c, err := w.Read(mustParse("foo://child.json"), target)
	require.NoError(t, err)

	require.Equal(t, &refChild{Name: "foo"}, c.Data())
	require.Equal(t, &refChild{}, target, "caller's value must not be filled")
}
