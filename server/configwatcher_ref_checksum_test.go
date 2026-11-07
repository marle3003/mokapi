package server_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/server"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// A change of a referenced file reloads the parent once per change, also when the
// content reverts to an earlier version. File changes arrive as Update events.
func TestConfigWatcher_ReferenceChecksum(t *testing.T) {
	logrus.SetOutput(io.Discard)

	setup := func(t *testing.T) (send func(evt dynamic.Event, checksum byte), parentReloads func() int, parentTime func() time.Time) {
		w := server.NewConfigWatcher(&static.Config{})
		mainPath := mustParse("foo://file.yml")
		refPath := mustParse("foo://ref.yml")
		mainConfig := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mainPath, Checksum: []byte{1}}}

		var ch chan dynamic.ConfigEvent
		w.AddProvider("foo", &testprovider{
			read: func(u *url.URL) (*dynamic.Config, error) {
				return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u, Checksum: []byte{1}}}, nil
			},
			start: func(configs chan dynamic.ConfigEvent, _ *safe.Pool) error {
				ch = configs
				return nil
			},
		})
		pool := safe.NewPool(context.Background())
		_ = w.Start(pool)
		t.Cleanup(pool.Stop)

		var m sync.Mutex
		var events []string
		var lastParentTime time.Time
		w.AddListener(func(e dynamic.ConfigEvent) {
			m.Lock()
			defer m.Unlock()
			events = append(events, e.Config.Info.Path())
			if e.Config == mainConfig {
				lastParentTime = e.Time
			}
		})
		parentTime = func() time.Time {
			m.Lock()
			defer m.Unlock()
			return lastParentTime
		}
		count := func() int {
			m.Lock()
			defer m.Unlock()
			n := 0
			for _, e := range events {
				if e == "foo://file.yml" {
					n++
				}
			}
			return n
		}

		ch <- dynamic.ConfigEvent{Name: mainPath.String(), Event: dynamic.Create, Config: mainConfig}
		c, err := w.Read(refPath, nil)
		require.NoError(t, err)
		dynamic.AddRef(mainConfig, c)
		send = func(evt dynamic.Event, checksum byte) {
			cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: refPath, Checksum: []byte{checksum}}}
			ch <- dynamic.ConfigEvent{Name: refPath.String(), Event: evt, Config: cfg}
			// no event is expected for unchanged content, so give the watcher time to settle
			time.Sleep(100 * time.Millisecond)
		}
		require.Eventually(t, func() bool { return count() == 1 }, time.Second, time.Millisecond)
		return send, count, parentTime
	}

	t.Run("change and revert of the referenced file reloads the parent each time", func(t *testing.T) {
		send, parentReloads, _ := setup(t)

		send(dynamic.Update, 2)
		require.Equal(t, 2, parentReloads(), "change 1 -> 2")
		send(dynamic.Update, 1)
		require.Equal(t, 3, parentReloads(), "change 2 -> 1")
	})

	t.Run("same change delivered twice reloads the parent once", func(t *testing.T) {
		send, parentReloads, _ := setup(t)

		send(dynamic.Update, 2)
		send(dynamic.Update, 2)
		require.Equal(t, 2, parentReloads(), "duplicate event with checksum 2")
	})

	t.Run("first Create of the referenced file with unchanged content does not reload the parent", func(t *testing.T) {
		send, parentReloads, _ := setup(t)

		send(dynamic.Create, 1)
		require.Equal(t, 1, parentReloads())
	})

	t.Run("reload of the parent takes over the time of the changed reference", func(t *testing.T) {
		send, parentReloads, parentTime := setup(t)
		before := time.Now()

		send(dynamic.Update, 2)

		require.Equal(t, 2, parentReloads())
		require.False(t, parentTime().Before(before), "parent time %v is older than the reference change", parentTime())
	})
}
