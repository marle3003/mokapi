package api

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
)

func TestHandler_Events(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "empty http events",
			fn: func(t *testing.T, h http.Handler) {
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
			fn: func(t *testing.T, h http.Handler) {
				events.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := events.Push("foo", events.NewTraits().WithNamespace("http"))
				event := events.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`[{"id":"%v","traits":{"namespace":"http"},"data":"foo","time":"%v"}]`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event",
			fn: func(t *testing.T, h http.Handler) {
				events.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := events.Push("foo", events.NewTraits().WithNamespace("http"))
				event := events.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+event.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"id":"%v","traits":{"namespace":"http"},"data":"foo","time":"%v"}`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event but not existing",
			fn: func(t *testing.T, h http.Handler) {
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
			defer events.Reset()

			cfg := &static.Config{}
			h := New(runtime.New(cfg), static.Api{})
			tc.fn(t, h)
		})
	}
}
