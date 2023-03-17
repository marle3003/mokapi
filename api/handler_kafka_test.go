package api

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Kafka(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler, app *runtime.App)
	}{
		{
			name: "/api/services",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": {
						Config: asyncapitest.NewConfig(asyncapitest.WithTitle("foo")),
					},
				},
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/services",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"name":"foo","version":"1.0","type":"kafka","topics":null}]`))
			},
		},
		{
			name: "/api/services/kafka/foo",
			app: &runtime.App{
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": {
						Config: asyncapitest.NewConfig(asyncapitest.WithTitle("foo")),
						Store:  store.New(asyncapitest.NewConfig(asyncapitest.WithTitle("foo")), enginetest.NewEngine()),
					},
				},
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				app.Monitor.Kafka.Messages.WithLabel("foo", "topic").Add(1)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/services/kafka/foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"name":"foo","description":"","version":"1.0","contact":null,"topics":null,"groups":[],"metrics":[{"name":"kafka_messages_total{service=\"foo\",topic=\"topic\"}","value":1}]}`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})
			tc.fn(t, h, tc.app)
		})
	}
}
