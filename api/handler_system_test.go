package api

import (
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_System(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "no event stores",
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(404))
			},
		},
		{
			name: "with event store",
			fn: func(t *testing.T, h http.Handler) {
				events.SetStore(1, events.NewTraits().WithNamespace("http"))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"traits":{"namespace":"http"},"size":1,"numEvents":0}]`))
			},
		},
		{
			name: "with namespace and name",
			fn: func(t *testing.T, h http.Handler) {
				events.SetStore(1, events.NewTraits().WithNamespace("kafka"))
				events.SetStore(1, events.NewTraits().WithNamespace("kafka").WithName("Kafka Testserver"))
				events.SetStore(1, events.NewTraits().WithNamespace("kafka").WithName("Kafka Testserver").With("topic", "foo"))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=kafka&name=Kafka Testserver",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"traits":{"namespace":"kafka"},"size":1,"numEvents":0},{"traits":{"name":"Kafka Testserver","namespace":"kafka"},"size":1,"numEvents":0}]`))
			},
		},
		{
			name: "with namespace, name and topic",
			fn: func(t *testing.T, h http.Handler) {
				events.SetStore(1, events.NewTraits().WithNamespace("kafka"))
				events.SetStore(1, events.NewTraits().WithNamespace("kafka").WithName("Kafka Testserver"))
				events.SetStore(1, events.NewTraits().WithNamespace("kafka").WithName("Kafka Testserver").With("topic", "foo"))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=kafka&name=Kafka Testserver&topic=foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"traits":{"namespace":"kafka"},"size":1,"numEvents":0},{"traits":{"name":"Kafka Testserver","namespace":"kafka"},"size":1,"numEvents":0},{"traits":{"name":"Kafka Testserver","namespace":"kafka","topic":"foo"},"size":1,"numEvents":0}]`))
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
