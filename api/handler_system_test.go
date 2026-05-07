package api

import (
	"mokapi/config/dynamic/dynamictest"
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
		fn   func(t *testing.T, h http.Handler, sm *events.StoreManager)
	}{
		{
			name: "default store with size 1",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasBody(`[{"traits":{},"size":1,"numEvents":0}]`),
				)
			},
		},
		{
			name: "with event store",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("foo"))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"traits":{},"size":1,"numEvents":0},{"traits":{"namespace":"foo"},"size":1,"numEvents":0}]`))
			},
		},
		{
			name: "with namespace and name",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("foo"))
				sm.SetStore(1, events.NewTraits().WithNamespace("foo").WithName("Kafka Testserver"))
				sm.SetStore(1, events.NewTraits().WithNamespace("foo").WithName("Kafka Testserver").With("topic", "foo"))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=foo&name=Kafka%20Testserver",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"traits":{},"size":1,"numEvents":0},{"traits":{"namespace":"foo"},"size":1,"numEvents":0},{"traits":{"name":"Kafka Testserver","namespace":"foo"},"size":1,"numEvents":0}]`))
			},
		},
		{
			name: "with namespace, name and topic",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("foo"))
				sm.SetStore(1, events.NewTraits().WithNamespace("foo").WithName("Kafka Testserver"))
				sm.SetStore(1, events.NewTraits().WithNamespace("foo").WithName("Kafka Testserver").With("topic", "foo"))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/system/events?namespace=foo&name=Kafka%20Testserver&topic=foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"traits":{},"size":1,"numEvents":0},{"traits":{"namespace":"foo"},"size":1,"numEvents":0},{"traits":{"name":"Kafka Testserver","namespace":"foo"},"size":1,"numEvents":0},{"traits":{"name":"Kafka Testserver","namespace":"foo","topic":"foo"},"size":1,"numEvents":0}]`))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{Event: static.Event{Store: map[string]static.Store{"default": {Size: 1}}}}
			app := runtime.New(cfg, &dynamictest.Reader{})

			h := New(app, static.Api{})
			tc.fn(t, h, app.Events)
		})
	}
}
