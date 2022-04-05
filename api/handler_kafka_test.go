package api

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Kafka(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "/api/services/kafka",
			app: &runtime.App{
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": {
						asyncapitest.NewConfig(asyncapitest.WithTitle("foo")),
					},
				},
			},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/services/kafka/foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"asyncapi":"2.0.0","info":{"title":"foo","version":"1.0"},"channels":null}`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, true)
			tc.fn(t, h)
		})
	}
}
