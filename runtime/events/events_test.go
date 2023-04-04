package events

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPush(t *testing.T) {
	testcase := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"no traits",
			func(t *testing.T) {
				err := Push(nil, NewTraits())
				require.EqualError(t, err, "empty traits not allowed")
			},
		},
		{
			"no store",
			func(t *testing.T) {
				err := Push(nil, NewTraits().WithNamespace("foo"))
				require.EqualError(t, err, "no store found for namespace=foo")
			},
		},
		{
			"no store matches",
			func(t *testing.T) {
				SetStore(10, NewTraits().WithNamespace("bar"))
				err := Push(nil, NewTraits().WithNamespace("foo"))
				require.EqualError(t, err, "no store found for namespace=foo")
			},
		},
		{
			"store matches",
			func(t *testing.T) {
				SetStore(10, NewTraits().WithNamespace("foo"))
				err := Push(nil, NewTraits().WithNamespace("foo"))
				require.NoError(t, err)
			},
		},
		{
			"store matches",
			func(t *testing.T) {
				SetStore(10, NewTraits().WithNamespace("foo"))
				err := Push(nil, NewTraits().WithNamespace("foo").WithName("bar"))
				require.NoError(t, err)
				err = Push(nil, NewTraits().WithNamespace("foo").WithName("foobar"))
				require.NoError(t, err)
				events := GetEvents(NewTraits().WithNamespace("foo"))
				require.Len(t, events, 2)
				events = GetEvents(NewTraits().WithNamespace("foo").WithName("bar"))
				require.Len(t, events, 1)
			},
		},
		{
			"store matches traits",
			func(t *testing.T) {
				SetStore(1, NewTraits().WithNamespace("foo"))
				SetStore(1, NewTraits().WithNamespace("foo").WithName("bar"))
				err := Push(nil, NewTraits().WithNamespace("foo").WithName("bar"))
				require.NoError(t, err)
				err = Push(nil, NewTraits().WithNamespace("foo").WithName("foobar"))
				require.NoError(t, err)
				events := GetEvents(NewTraits().WithNamespace("foo"))
				require.Len(t, events, 2)
				events = GetEvents(NewTraits().WithNamespace("foo").WithName("bar"))
				require.Len(t, events, 1)
			},
		},
		{
			"get all events",
			func(t *testing.T) {
				SetStore(10, NewTraits().WithNamespace("foo"))
				err := Push(nil, NewTraits().WithNamespace("foo"))
				require.NoError(t, err)

				events := GetEvents(NewTraits())
				require.Len(t, events, 1)
			},
		},
		{
			"get events by namespace",
			func(t *testing.T) {
				SetStore(10, NewTraits().WithNamespace("foo"))
				err := Push(nil, NewTraits().WithNamespace("foo"))
				require.NoError(t, err)

				events := GetEvents(NewTraits().WithNamespace("foo"))
				require.Len(t, events, 1)
			},
		},
		{
			"reset store with traits",
			func(t *testing.T) {
				SetStore(10, NewTraits().WithNamespace("foo"))
				SetStore(10, NewTraits().WithNamespace("bar"))
				SetStore(10, NewTraits().WithNamespace("foo").WithName("name"))
				ResetStores(NewTraits().WithNamespace("foo"))

				require.Len(t, stores, 1)
			},
		},
		{
			"Clean up, ensure size and oldest is removed",
			func(t *testing.T) {
				SetStore(2, NewTraits().WithNamespace("foo"))
				err := Push(1, NewTraits().WithNamespace("foo"))
				require.NoError(t, err)
				err = Push(2, NewTraits().WithNamespace("foo"))
				require.NoError(t, err)
				err = Push(3, NewTraits().WithNamespace("foo"))
				require.NoError(t, err)

				events := GetEvents(NewTraits().WithNamespace("foo"))
				require.Len(t, events, 2)
				require.Equal(t, 3, events[0].Data)
				require.Equal(t, 2, events[1].Data)
			},
		},
	}

	for _, tc := range testcase {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer Reset()
			tc.f(t)
		})
	}
}
