package events_test

import (
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	testcase := []struct {
		name string
		f    func(t *testing.T, sm *events.StoreManager)
	}{
		{
			"no traits",
			func(t *testing.T, sm *events.StoreManager) {
				err := sm.Push(nil, events.NewTraits())
				require.EqualError(t, err, "empty traits not allowed")
			},
		},
		{
			"no store",
			func(t *testing.T, sm *events.StoreManager) {
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo"))
				require.EqualError(t, err, "no store found for namespace=foo")
			},
		},
		{
			"no store matches",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("bar"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo"))
				require.EqualError(t, err, "no store found for namespace=foo")
			},
		},
		{
			"store matches",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("foo"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo"))
				require.NoError(t, err)
			},
		},
		{
			"store with empty traits",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits())
				err := sm.Push(nil, events.NewTraits().With("level", "debug"))
				require.NoError(t, err)
			},
		},
		{
			"store matches",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("foo"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo").WithName("bar"))
				require.NoError(t, err)
				err = sm.Push(nil, events.NewTraits().WithNamespace("foo").WithName("foobar"))
				require.NoError(t, err)
				evts := sm.GetEvents(events.NewTraits().WithNamespace("foo"))
				require.Len(t, evts, 2)
				evts = sm.GetEvents(events.NewTraits().WithNamespace("foo").WithName("bar"))
				require.Len(t, evts, 1)
			},
		},
		{
			"store matches traits",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("foo"))
				sm.SetStore(1, events.NewTraits().WithNamespace("foo").WithName("bar"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo").WithName("bar"))
				require.NoError(t, err)
				err = sm.Push(nil, events.NewTraits().WithNamespace("foo").WithName("foobar"))
				require.NoError(t, err)
				evts := sm.GetEvents(events.NewTraits().WithNamespace("foo"))
				require.Len(t, evts, 2)
				evts = sm.GetEvents(events.NewTraits().WithNamespace("foo").WithName("bar"))
				require.Len(t, evts, 1)
			},
		},
		{
			"get all events",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("foo"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo"))
				require.NoError(t, err)

				evts := sm.GetEvents(events.NewTraits())
				require.Len(t, evts, 1)
			},
		},
		{
			"get events by namespace",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("foo"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo"))
				require.NoError(t, err)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("foo"))
				require.Len(t, evts, 1)
			},
		},
		{
			"reset store with traits",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("foo"))
				sm.SetStore(10, events.NewTraits().WithNamespace("bar"))
				sm.SetStore(10, events.NewTraits().WithNamespace("foo").WithName("name"))
				sm.ResetStores(events.NewTraits().WithNamespace("foo"))

				require.Len(t, sm.GetStores(events.NewTraits().WithNamespace("foo")), 0)
				require.Len(t, sm.GetStores(events.NewTraits().WithNamespace("bar")), 1)
			},
		},
		{
			"Clean up, ensure size and oldest is removed",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("foo"))
				err := sm.Push(&eventstest.Event{Name: "1"}, events.NewTraits().WithNamespace("foo"))
				require.NoError(t, err)
				err = sm.Push(&eventstest.Event{Name: "2"}, events.NewTraits().WithNamespace("foo"))
				require.NoError(t, err)
				err = sm.Push(&eventstest.Event{Name: "3"}, events.NewTraits().WithNamespace("foo"))
				require.NoError(t, err)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("foo"))
				require.Len(t, evts, 2)
				require.Equal(t, "3", evts[0].Data.Title())
				require.Equal(t, "2", evts[1].Data.Title())
			},
		},
		{
			"get events by traits",
			func(t *testing.T, sm *events.StoreManager) {
				sm.SetStore(10, events.NewTraits().WithNamespace("foo").With("name", "bar"))
				err := sm.Push(nil, events.NewTraits().WithNamespace("foo").With("name", "bar").With("type", "baz").With("client", "client-1"))
				require.NoError(t, err)

				result := sm.GetEvents(events.NewTraits().WithNamespace("foo").With("name", "bar").With("type", "baz"))
				require.Len(t, result, 1)

				result = sm.GetEvents(events.NewTraits().WithNamespace("foo").With("type", "baz"))
				require.Len(t, result, 1)

				result = sm.GetEvents(events.NewTraits().WithNamespace("foo").With("type", "yuh"))
				require.Len(t, result, 0)
			},
		},
	}

	for _, tc := range testcase {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t, events.NewStoreManager(&index{}))
		})
	}
}

type index struct{}

func (*index) Add(string, any) {}

func (*index) Delete(string) {}

// TestStoreManager_ConcurrentAccess mimics parallel config updates, which add and reset stores
// while other goroutines read them.
func TestStoreManager_ConcurrentAccess(t *testing.T) {
	sm := events.NewStoreManager(nil)

	operations := []func(){
		func() { sm.SetStore(10, events.NewTraits().WithNamespace("http").WithName("foo")) },
		func() { sm.ResetStores(events.NewTraits().WithNamespace("http").WithName("foo")) },
		func() { sm.GetStores(events.NewTraits().WithNamespace("http")) },
	}

	var wg sync.WaitGroup
	for _, op := range operations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				op()
			}
		}()
	}
	wg.Wait()
}

func TestPushEvent(t *testing.T) {
	t.Run("returns the stored event", func(t *testing.T) {
		sm := events.NewStoreManager(&index{})
		traits := events.NewTraits().WithNamespace("foo")
		sm.SetStore(10, traits)

		evt, err := sm.PushEvent(&eventstest.Event{Name: "1"}, traits)

		require.NoError(t, err)
		require.True(t, evt.IsValid())
		require.False(t, evt.Time.IsZero())
		require.Equal(t, "1", evt.Data.Title())
		stored := sm.GetEvents(traits)
		require.Len(t, stored, 1)
		require.Equal(t, evt.Id, stored[0].Id)
		require.Equal(t, evt.Id, sm.GetEvent(evt.Id).Id)
	})

	t.Run("returns the error and no event", func(t *testing.T) {
		sm := events.NewStoreManager(&index{})

		evt, err := sm.PushEvent(nil, events.NewTraits().WithNamespace("foo"))

		require.EqualError(t, err, "no store found for namespace=foo")
		require.False(t, evt.IsValid())
	})
}

// recordingIndex records index writes so tests can tell whether an event was (re-)indexed.
type recordingIndex struct {
	m       sync.Mutex
	added   []indexed
	deleted []string
}

type indexed struct {
	id   string
	data map[string]any
}

func (i *recordingIndex) Add(id string, data any) {
	i.m.Lock()
	defer i.m.Unlock()
	i.added = append(i.added, indexed{id: id, data: data.(map[string]any)})
}

func (i *recordingIndex) Delete(id string) {
	i.m.Lock()
	defer i.m.Unlock()
	i.deleted = append(i.deleted, id)
}

func (i *recordingIndex) adds() []indexed {
	i.m.Lock()
	defer i.m.Unlock()
	return append([]indexed(nil), i.added...)
}

func TestUpdate(t *testing.T) {
	t.Run("re-indexes the stored event with the changed fields", func(t *testing.T) {
		idx := &recordingIndex{}
		sm := events.NewStoreManager(idx)
		traits := events.NewTraits().WithNamespace("foo")
		sm.SetStore(10, traits)
		data := &eventstest.Event{Name: "before"}
		evt, err := sm.PushEvent(data, traits)
		require.NoError(t, err)

		data.Name = "after"
		sm.Update(data)

		adds := idx.adds()
		require.Len(t, adds, 2)
		require.Equal(t, evt.Id, adds[1].id)
		require.Equal(t, "after", adds[1].data["name"])
	})

	t.Run("data that was never pushed is not indexed", func(t *testing.T) {
		idx := &recordingIndex{}
		sm := events.NewStoreManager(idx)
		traits := events.NewTraits().WithNamespace("foo")
		sm.SetStore(10, traits)
		_, err := sm.PushEvent(&eventstest.Event{Name: "pushed"}, traits)
		require.NoError(t, err)

		sm.Update(&eventstest.Event{Name: "unknown"})

		require.Len(t, idx.adds(), 1)
	})

	t.Run("an evicted event is not indexed again", func(t *testing.T) {
		idx := &recordingIndex{}
		sm := events.NewStoreManager(idx)
		traits := events.NewTraits().WithNamespace("foo")
		sm.SetStore(1, traits)
		evicted := &eventstest.Event{Name: "evicted"}
		_, err := sm.PushEvent(evicted, traits)
		require.NoError(t, err)
		_, err = sm.PushEvent(&eventstest.Event{Name: "newer"}, traits)
		require.NoError(t, err)

		sm.Update(evicted)

		require.Len(t, idx.adds(), 2)
	})

	t.Run("without index nothing happens", func(t *testing.T) {
		sm := events.NewStoreManager(nil)
		traits := events.NewTraits().WithNamespace("foo")
		sm.SetStore(10, traits)
		data := &eventstest.Event{Name: "foo"}
		_, err := sm.PushEvent(data, traits)
		require.NoError(t, err)

		sm.Update(data)
	})
}

// TestUpdate_ConcurrentWithPush runs under -race: Update looks into the stores while Push
// adds and evicts events.
func TestUpdate_ConcurrentWithPush(t *testing.T) {
	sm := events.NewStoreManager(&recordingIndex{})
	traits := events.NewTraits().WithNamespace("foo")
	sm.SetStore(2, traits)
	data := &eventstest.Event{Name: "first"}
	_, err := sm.PushEvent(data, traits)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 50 {
				_ = sm.Push(&eventstest.Event{Name: "other"}, traits)
			}
		}()
		go func() {
			defer wg.Done()
			for range 50 {
				sm.Update(data)
			}
		}()
	}
	wg.Wait()
}
