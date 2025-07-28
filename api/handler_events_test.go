package api

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
)

func TestHandler_Events(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, h http.Handler, sm *events.StoreManager)
	}{
		{
			name: "empty http events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[]`))
			},
		},
		{
			name: "with http events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("http"))
				event := sm.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`[{"id":"%v","traits":{"namespace":"http"},"data":{"Name":"foo","api":""},"time":"%v"}]`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("http"))
				event := sm.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+event.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"id":"%v","traits":{"namespace":"http"},"data":{"Name":"foo","api":""},"time":"%v"}`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event but not existing",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/1234",
					nil,
					"",
					h,
					try.HasStatusCode(404))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg)

			h := New(app, static.Api{})
			tc.fn(t, h, app.Events)
		})
	}
}
