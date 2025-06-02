package server

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/schema"
	"mokapi/try"
	"testing"
	"time"
)

func TestMqttServer(t *testing.T) {
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	c := asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("mqtt12", "mqtt", addr),
		asyncapi3test.WithChannel("foo",
			asyncapi3test.WithMessage("foo",
				asyncapi3test.WithPayload(
					&schema.Schema{Type: schema.Types{"string"}},
				),
			),
		),
	)

	cfg := &static.Config{}
	m := NewMqttManager(nil, runtime.New(cfg))
	defer m.Stop()
	m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}, Data: c}})

	// wait for kafka start
	time.Sleep(500 * time.Millisecond)

	require.Len(t, m.clusters, 1)
	_, ok := m.clusters["foo"]
	require.True(t, ok, "cluster exists")
}
