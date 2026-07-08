package api_test

import (
	"fmt"
	"io"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/runtime"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandler_Websocket(t *testing.T) {
	logrus.SetOutput(io.Discard)

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
					runtimetest.WithWebsocket(
						asyncapi3test.NewConfig(
							asyncapi3test.WithInfo("foo", "bar", "1.0"),
						),
					),
				)
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"websocket","metrics":{"websocket_messages_total":0,"websocket_message_timestamp":0}}]`,
		},
		{
			name: "get Websocket services",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})
				_, _ = app.Websocket.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("websocket.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "websocket", "1.0"),
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
			requestUrl:   "http://foo.api/api/services/websocket",
			responseBody: `[{"name":"foo","description":"websocket","contact":{"name":"mokapi","url":"https://mokapi.io","email":"info@mokapi.io"},"version":"1.0","type":"websocket","metrics":{"websocket_messages_total":0,"websocket_message_timestamp":0}}]`,
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

				_, _ = app.Websocket.Add(cfg, enginetest.NewEngine())
				return app
			},
			requestUrl:   "http://foo.api/api/services/websocket/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"mokapi","host":":80","protocol":"ws","title":"Mokapi Default Server","summary":"Automatically added server because no servers are defined in the AsyncAPI spec","description":""}],"configs":[{"id":"64613435-3062-6462-3033-316532633233","url":"file://foo.yml","provider":"test","time":"2023-12-27T13:01:30Z"}]}`,
		},
		{
			name: "channel with parameter",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})
				addr := fmt.Sprintf(":%v", try.GetFreePort())
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("foo", "ws", addr),
						asyncapi3test.WithChannel("chats/{chatId}",
							asyncapi3test.WithParameter("chatId", &asyncapi3.Parameter{}),
						),
					),
				}
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")

				mi, err := app.Websocket.Add(cfg, enginetest.NewEngine())
				require.NoError(t, err)
				mi.Store.Channels = map[string]*websocket.Channel{
					"chats/1234z": {Name: "chats/1234z"},
				}

				return app
			},
			requestUrl:   "http://foo.api/api/services/websocket/foo/channels",
			responseBody: `[{"name":"chats/{chatId}","instances":[{"name":"chats/1234z","parameters":{"chatId":"1234z"}}],"metrics":{"websocket_messages_total":0,"websocket_message_timestamp":0}}]`,
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

func TestHandler_Websocket_Metrics(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
		addMetrics   func(monitor *monitor.Monitor)
	}{
		{
			name:         "service list with metric",
			app:          runtimetest.NewApp(runtimetest.WithWebsocketInfo("foo", getWebsocketInfo(asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"))))),
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","version":"1.0","type":"websocket","metrics":{"websocket_messages_total":1,"websocket_message_timestamp":12345678}}]`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Websocket.Messages.WithLabel("foo", "channel-1").Add(1)
				monitor.Websocket.LastMessage.WithLabel("foo", "channel-1").Set(12345678)
			},
		},
		{
			name: "cluster with metric",
			app: runtimetest.NewApp(
				runtimetest.WithWebsocketInfo("foo", getWebsocketInfo(
					asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"),
						asyncapi3test.WithChannel("foo"),
					),
				))),
			requestUrl:   "http://foo.api/api/services/websocket/foo",
			responseBody: `{"name":"foo","version":"1.0","channels":[{"name":"foo","metrics":{"websocket_messages_total":1,"websocket_message_timestamp":12345678}}]}`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Websocket.Messages.WithLabel("foo", "channel-1").Add(1)
				monitor.Websocket.LastMessage.WithLabel("foo", "channel-1").Set(12345678)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := api.New(tc.app, static.Api{})
			tc.addMetrics(tc.app.Monitor)

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

func getWebsocketInfo(config *asyncapi3.Config) *runtime.WebsocketInfo {
	return &runtime.WebsocketInfo{
		Config: config,
		Store:  websocket.New(config, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewWebsocket()),
	}
}
