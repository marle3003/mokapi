package dynamic_test

import (
	"mokapi/config/dynamic"
	"net/url"
	"sync"
	"testing"
)

// Reproduces the data race of the AddRef listener on the reference's ConfigInfo:
// the watcher goroutine calls Info.Update for the next change while the listener of
// the previous change (started with "go l(e)" in invokeListeners) still runs. The
// listener must only use the snapshot in the event. Only meaningful with -race.
func TestAddRef_ListenerRacesWithInfoUpdate(t *testing.T) {
	parent := &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "foo", Host: "parent.yml"}}}
	ref := &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "foo", Host: "ref.yml"}, Checksum: []byte{1}}}
	dynamic.AddRef(parent, ref)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		// watcher goroutine: addOrUpdate updates the reference ...
		ref.Info.Update([]byte{byte(i)})
		// ... and invokeListeners runs the AddRef listener in a new goroutine with a snapshot
		e := dynamic.ConfigEvent{Event: dynamic.Update, Config: ref, Checksum: ref.Info.Checksum, Time: ref.Info.Time}
		for _, l := range ref.Listeners.List() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				l(e)
			}()
		}
	}
	wg.Wait()
}
