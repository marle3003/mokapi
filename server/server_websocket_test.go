package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/schema"
	"mokapi/server"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestWebsocketServer(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, m *server.WebsocketManager)
	}{
		{
			name: "TestWebsocketServer",
			test: func(t *testing.T, m *server.WebsocketManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				c := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("ws12", "ws", addr),
					asyncapi3test.WithChannel("/foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
					),
				)

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}, Data: c}})

				// wait for websocket start
				time.Sleep(500 * time.Millisecond)

				client := http.Client{}
				res, err := client.Get("http://" + addr + "/foo")
				require.NoError(t, err)
				require.Equal(t, http.StatusUpgradeRequired, res.StatusCode)
			},
		},
		{
			name: "kafka topic should not be available",
			test: func(t *testing.T, m *server.WebsocketManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("ws12", "ws", addr),
					asyncapi3test.WithServer("kafka", "kafka", fmt.Sprintf("127.0.0.1:%v", try.GetFreePort())),
					asyncapi3test.WithChannel("/foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
						asyncapi3test.AssignToServer("#/servers/ws12"),
					),
					asyncapi3test.WithChannel("/bar",
						asyncapi3test.WithMessage("bar",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
						asyncapi3test.AssignToServer("#/servers/kafka"),
					),
				)
				c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}, Data: cfg}
				err := cfg.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)

				m.UpdateConfig(dynamic.ConfigEvent{Config: c})

				// wait for websocket start
				time.Sleep(500 * time.Millisecond)

				client := http.Client{}
				res, err := client.Get("http://" + addr + "/foo")
				require.NoError(t, err)
				require.Equal(t, http.StatusUpgradeRequired, res.StatusCode)

				res, err = client.Get("http://" + addr + "/bar")
				require.NoError(t, err)
				require.Equal(t, http.StatusNotFound, res.StatusCode)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := server.NewWebsocketManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
			defer m.Stop()

			tc.test(t, m)
		})
	}
}
