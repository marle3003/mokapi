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
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// A single file with empty entries (`foo:` without value) next to a valid channel: the
// dashboard API serves the valid channel and ignores the empty entries.
func TestHandler_AsyncApi_FileWithEmptyEntries(t *testing.T) {
	logrus.SetOutput(io.Discard)

	spec := func(protocol string) string {
		return fmt.Sprintf(`
asyncapi: 3.0.0
info: {title: foo, version: '1.0'}
servers:
  empty:
  dev: {host: 'localhost:%d', protocol: %s}
channels:
  empty:
  broken:
    address: broken
    servers: [null]
    parameters: {id: }
    messages: {created: }
  orders:
    address: orders
    messages:
      empty:
      order: {payload: {type: string}}
operations:
  empty:
  noChannel: {action: send}
  send:
    action: send
    channel: {$ref: '#/channels/orders'}
    messages: [null]
`, try.GetFreePort(), protocol)
	}

	testcases := []struct {
		name     string
		protocol string
		add      func(app *runtime.App, c *dynamic.Config) error
		urls     []string
	}{
		{
			name:     "kafka",
			protocol: "kafka",
			add: func(app *runtime.App, c *dynamic.Config) error {
				_, err := app.Kafka.Add(c, enginetest.NewEngine())
				return err
			},
			urls: []string{
				"/api/services/kafka/foo",
				"/api/services/kafka/foo/topics",
				"/api/services/kafka/foo/topics/orders",
				"/api/services/kafka/foo/channels/orders",
				"/api/services/kafka/foo/asyncapi.json",
			},
		},
		{
			name:     "mqtt",
			protocol: "mqtt",
			add: func(app *runtime.App, c *dynamic.Config) error {
				_, err := app.Mqtt.Add(c, enginetest.NewEngine())
				return err
			},
			urls: []string{
				"/api/services/mqtt/foo",
				"/api/services/mqtt/foo/topics",
			},
		},
		{
			name:     "websocket",
			protocol: "ws",
			add: func(app *runtime.App, c *dynamic.Config) error {
				_, err := app.Websocket.Add(c, enginetest.NewEngine())
				return err
			},
			urls: []string{
				"/api/services/websocket/foo",
				"/api/services/websocket/foo/channels",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *asyncapi3.Config
			require.NoError(t, yaml.Unmarshal([]byte(spec(tc.protocol)), &cfg))
			c := &dynamic.Config{Info: dynamictest.NewConfigInfo(), Data: cfg}
			require.NoError(t, cfg.Parse(c, &dynamictest.Reader{}))

			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			require.NoError(t, tc.add(app, c))
			h := api.New(app, static.Api{})

			for _, u := range tc.urls {
				t.Run(u, func(t *testing.T) {
					require.NotPanics(t, func() {
						try.Handler(t, http.MethodGet, "http://foo.api"+u, nil, "", h,
							try.HasStatusCode(http.StatusOK),
							try.BodyContains("orders"))
					})
				})
			}
		})
	}
}
