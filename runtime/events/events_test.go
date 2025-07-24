package events_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"testing"
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
	}

	for _, tc := range testcase {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t, &events.StoreManager{})
		})
	}
}
