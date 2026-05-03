package api_test

import (
	"fmt"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandler_Mqtt(t *testing.T) {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	testcases := []struct {
		name         string
		app          func() *runtime.App
		requestUrl   string
		responseBody string
	}{
		{
			name: "get services",
			app: func() *runtime.App {
				return runtimetest.NewApp(
					runtimetest.WithMqtt(
						asyncapi3test.NewConfig(
							asyncapi3test.WithInfo("foo", "bar", "1.0"),
						),
					),
				)
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"mqtt"}]`,
		},
		{
			name: "get MQTT services",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})
				_, _ = app.Mqtt.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("mqtt.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "mqtt", "1.0"),
						asyncapi3test.WithContact("mokapi", "https://mokapi.io", "info@mokapi.io"),
					),
				}, enginetest.NewEngine())
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "kafka", "1.0"),
						asyncapi3test.WithContact("mokapi", "https://mokapi.io", "info@mokapi.io"),
					),
				}, enginetest.NewEngine())
				return app
			},
			requestUrl:   "http://foo.api/api/services/mqtt",
			responseBody: `[{"name":"foo","description":"mqtt","contact":{"name":"mokapi","url":"https://mokapi.io","email":"info@mokapi.io"},"version":"1.0"}]`,
		},
		{
			name: "get specific",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
					),
				}
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")

				_, _ = app.Mqtt.Add(cfg, enginetest.NewEngine())
				return app
			},
			requestUrl:   "http://foo.api/api/services/mqtt/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"mokapi","host":":1883","protocol":"mqtt","title":"Mokapi Default Broker","summary":"Automatically added broker because no servers are defined in the AsyncAPI spec","description":""}],"configs":[{"id":"64613435-3062-6462-3033-316532633233","url":"file://foo.yml","provider":"test","time":"2023-12-27T13:01:30Z"}]}`,
		},
		{
			name: "topic with parameter",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})
				addr := fmt.Sprintf(":%v", try.GetFreePort())
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("foo", "mqtt", addr),
						asyncapi3test.WithChannel("sensors/{sensorId}/data",
							asyncapi3test.WithParameter("sensorId", &asyncapi3.Parameter{}),
						),
					),
				}
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")

				mi, err := app.Mqtt.Add(cfg, enginetest.NewEngine())
				require.NoError(t, err)
				mi.Topics["sensors/1234z/data"] = &store.Topic{Name: "sensors/1234z/data"}

				return app
			},
			requestUrl:   "http://foo.api/api/services/mqtt/foo/topics",
			responseBody: `[{"name":"sensors/{sensorId}/data","description":"","messages":null,"instances":[{"name":"sensors/1234z/data","parameters":{"sensorId":"1234z"}}]}]`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := api.New(tc.app(), static.Api{})

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(tc.responseBody))
		})
	}
}
