package events_test

import (
	"context"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/search"
	"mokapi/safe"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIndex_Http(t *testing.T) {

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "Search by name",
			test: func(t *testing.T, app *runtime.App) {
				trait := events.NewTraits().WithNamespace("test")

				app.Events.SetStore(10, trait)
				err := app.Events.Push(&eventstest.Event{
					Name: "foo",
					Api:  "My API",
				}, trait)
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "My API", r.Results[0].Domain)
				require.Equal(t, "foo", r.Results[0].Title)
				require.Equal(t, []string{"<mark>foo</mark>"}, r.Results[0].Fragments)
				require.Equal(t, "test", r.Results[0].Params["namespace"])
				require.NotEmpty(t, r.Results[0].Params["id"])
			},
		},
		{
			name: "Search by type",
			test: func(t *testing.T, app *runtime.App) {
				trait := events.NewTraits().WithNamespace("test")

				app.Events.SetStore(10, trait)
				err := app.Events.Push(&eventstest.Event{
					Name: "foo",
					Api:  "My API",
				}, trait)
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "type:event", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "My API", r.Results[0].Domain)
				require.Equal(t, "foo", r.Results[0].Title)
				require.Equal(t, []string{"<mark>event</mark>"}, r.Results[0].Fragments)
				require.Equal(t, "test", r.Results[0].Params["namespace"])
				require.NotEmpty(t, r.Results[0].Params["id"])
			},
		},
		{
			name: "Search by type without value",
			test: func(t *testing.T, app *runtime.App) {
				trait := events.NewTraits().WithNamespace("test")

				app.Events.SetStore(10, trait)
				err := app.Events.Push(&eventstest.Event{
					Name: "foo",
					Api:  "My API",
				}, trait)
				require.NoError(t, err)
				_, err = app.Search(search.Request{QueryText: "type:", Limit: 10})
				require.Error(t, err)
			},
		},
		{
			name: "when event is removed, it is also removed from index",
			test: func(t *testing.T, app *runtime.App) {
				trait := events.NewTraits().WithNamespace("test")

				app.Events.SetStore(1, trait)
				err := app.Events.Push(&eventstest.Event{
					Name: "foo",
					Api:  "My API",
				}, trait)
				require.NoError(t, err)

				err = app.Events.Push(&eventstest.Event{
					Name: "bar",
					Api:  "My API",
				}, trait)
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "type:event", Limit: 10})
					require.NoError(t, err)
					if len(r.Results) == 0 {
						return false
					}
					return r.Results[0].Title == "bar"
				})
				require.Len(t, r.Results, 1)

				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "My API", r.Results[0].Domain)
				require.Equal(t, "bar", r.Results[0].Title)
				require.Equal(t, []string{"<mark>event</mark>"}, r.Results[0].Fragments)
				require.Equal(t, "test", r.Results[0].Params["namespace"])
				require.NotEmpty(t, r.Results[0].Params["id"])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled:  true,
							InMemory: true,
						},
					},
				})

			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			tc.test(t, app)
		})
	}
}

func waitSearchIndex(t *testing.T, check func() bool) {
	deadline := time.Now().Add(2 * time.Second)

	for {
		if check() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("wait search index reached deadline")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
