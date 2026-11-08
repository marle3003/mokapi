package server_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/try"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Two files with the same AsyncAPI title are merged into one config.
// Deleting one of them must remove only its contribution.
type mergedFiles struct {
	portA, portB int
	a, b         dynamic.ConfigEvent
}

func newMergedFiles(protocol string) mergedFiles {
	f := mergedFiles{portA: try.GetFreePort(), portB: try.GetFreePort()}
	f.a = mergedFileEvent("a.yml", asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("a", protocol, fmt.Sprintf("127.0.0.1:%d", f.portA)),
		asyncapi3test.WithChannel("fromA"),
	))
	f.b = mergedFileEvent("b.yml", asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("b", protocol, fmt.Sprintf("127.0.0.1:%d", f.portB)),
		asyncapi3test.WithChannel("fromB"),
	))
	return f
}

func mergedFileEvent(file string, cfg *asyncapi3.Config) dynamic.ConfigEvent {
	return dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl(file)}, Data: cfg}}
}

func (f mergedFiles) deleteB() dynamic.ConfigEvent {
	e := f.b
	e.Event = dynamic.Delete
	return e
}

func (f mergedFiles) requireOnlyFileAIsLeft(t *testing.T, merged *asyncapi3.Config, urls []*dynamic.Config) {
	t.Helper()
	require.NotNil(t, merged)
	require.Len(t, urls, 1, "only file a.yml must remain in the store")
	require.Equal(t, "a.yml", urls[0].Info.Url.String())
	_, hasServerB := merged.Servers.Get("b")
	require.False(t, hasServerB, "server of deleted file must not be in merged config")
	require.NotContains(t, merged.Channels, "fromB", "channel of deleted file must not be in merged config")
	require.Contains(t, merged.Channels, "fromA")
	requirePortOpen(t, f.portA)
	requirePortClosed(t, f.portB)
}

func requirePortOpen(t *testing.T, port int) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	require.NoError(t, err, "port %d must still be open", port)
	_ = conn.Close()
}

func TestKafkaManager_DeleteOneOfTwoFilesWithSameTitle(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewKafkaManager(nil, app)
	defer m.Stop()

	f := newMergedFiles("kafka")
	m.UpdateConfig(f.a)
	m.UpdateConfig(f.b)
	requirePortOpen(t, f.portB)

	m.UpdateConfig(f.deleteB())

	info := app.Kafka.Get("foo")
	require.NotNil(t, info)
	f.requireOnlyFileAIsLeft(t, info.Config, info.Configs())
}

func TestMqttManager_DeleteOneOfTwoFilesWithSameTitle(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewMqttManager(nil, app)
	defer m.Stop()

	f := newMergedFiles("mqtt")
	m.UpdateConfig(f.a)
	m.UpdateConfig(f.b)
	requirePortOpen(t, f.portB)

	m.UpdateConfig(f.deleteB())

	info := app.Mqtt.Get("foo")
	require.NotNil(t, info)
	f.requireOnlyFileAIsLeft(t, info.Config, info.Configs())
}

func TestWebsocketManager_DeleteOneOfTwoFilesWithSameTitle(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewWebsocketManager(nil, app)
	defer m.Stop()

	f := newMergedFiles("ws")
	m.UpdateConfig(f.a)
	m.UpdateConfig(f.b)
	requirePortOpen(t, f.portB)

	m.UpdateConfig(f.deleteB())

	info := app.Websocket.Get("foo")
	require.NotNil(t, info)
	f.requireOnlyFileAIsLeft(t, info.Config, info.Configs())
}
