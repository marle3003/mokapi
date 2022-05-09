package api

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/try"
	"net/http"
	"testing"
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
					"http://foo.api/api/events/http",
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
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer events.Reset()
			h := New(runtime.New(), static.Api{})
			tc.fn(t, h)
		})
	}
}
