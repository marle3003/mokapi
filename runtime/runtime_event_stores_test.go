package runtime_test

import (
	"io"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// testEvent is a minimal event payload.
type testEvent struct{}

func (*testEvent) Title() string { return "" }

func newEventStoreApp() *runtime.App {
	logrus.SetOutput(io.Discard)
	cfg := &static.Config{}
	cfg.Event.Store = map[string]static.Store{"default": {Size: 10}}
	return runtime.New(cfg, &dynamictest.Reader{})
}

func serviceTraits(namespace string) events.Traits {
	return events.NewTraits().WithNamespace(namespace).WithName("foo")
}

// serviceStores returns the stores dedicated to the service "foo" of the namespace. Parent stores
// of the namespace and the default store are not part of it.
func serviceStores(app *runtime.App, namespace string) []events.StoreInfo {
	var result []events.StoreInfo
	for _, s := range app.Events.GetStores(serviceTraits(namespace)) {
		if s.Traits.GetName() == "foo" {
			result = append(result, s)
		}
	}
	return result
}

func TestEventStores_RemovingLastFileResetsStores(t *testing.T) {
	t.Run("mqtt", func(t *testing.T) {
		app := newEventStoreApp()
		file := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		_, err := app.Mqtt.Add(file, enginetest.NewEngine())
		require.NoError(t, err)
		require.NoError(t, app.Events.Push(&testEvent{}, serviceTraits("mqtt").With("topic", "bar")))
		require.Len(t, app.Events.GetEvents(serviceTraits("mqtt")), 1)

		app.Mqtt.Remove(file)

		require.Empty(t, serviceStores(app, "mqtt"))
		require.Empty(t, app.Events.GetEvents(serviceTraits("mqtt")))
	})

	t.Run("websocket", func(t *testing.T) {
		app := newEventStoreApp()
		file := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		_, err := app.Websocket.Add(file, enginetest.NewEngine())
		require.NoError(t, err)
		require.NoError(t, app.Events.Push(&testEvent{}, serviceTraits("websocket").With("channel", "bar")))
		require.Len(t, app.Events.GetEvents(serviceTraits("websocket")), 1)

		app.Websocket.Remove(file)

		require.Empty(t, serviceStores(app, "websocket"))
		require.Empty(t, app.Events.GetEvents(serviceTraits("websocket")))
	})
}

func TestEventStores_WebsocketAddCreatesOwnStore(t *testing.T) {
	app := newEventStoreApp()
	file := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))

	_, err := app.Websocket.Add(file, enginetest.NewEngine())
	require.NoError(t, err)

	require.Len(t, serviceStores(app, "websocket"), 1)
	require.Empty(t, serviceStores(app, "mqtt"))
}

func TestEventStores_WebsocketAddKeepsMqttStoresOfSameTitle(t *testing.T) {
	app := newEventStoreApp()
	mqttFile := getConfig("https://mokapi.io/mqtt.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
	wsFile := getConfig("https://mokapi.io/ws.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
	_, err := app.Mqtt.Add(mqttFile, enginetest.NewEngine())
	require.NoError(t, err)
	require.NoError(t, app.Events.Push(&testEvent{}, serviceTraits("mqtt").With("topic", "bar")))

	_, err = app.Websocket.Add(wsFile, enginetest.NewEngine())
	require.NoError(t, err)

	require.Len(t, app.Events.GetEvents(serviceTraits("mqtt")), 1)
}

func TestEventStores_RemovingOneOfTwoFilesKeepsStores(t *testing.T) {
	t.Run("mqtt", func(t *testing.T) {
		app := newEventStoreApp()
		a := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		b := getConfig("https://mokapi.io/b.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		_, err := app.Mqtt.Add(a, enginetest.NewEngine())
		require.NoError(t, err)
		_, err = app.Mqtt.Add(b, enginetest.NewEngine())
		require.NoError(t, err)
		require.NoError(t, app.Events.Push(&testEvent{}, serviceTraits("mqtt").With("topic", "bar")))

		app.Mqtt.Remove(a)

		require.Len(t, app.Events.GetEvents(serviceTraits("mqtt")), 1)
	})

	t.Run("websocket", func(t *testing.T) {
		app := newEventStoreApp()
		a := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		b := getConfig("https://mokapi.io/b.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		_, err := app.Websocket.Add(a, enginetest.NewEngine())
		require.NoError(t, err)
		_, err = app.Websocket.Add(b, enginetest.NewEngine())
		require.NoError(t, err)
		require.NoError(t, app.Events.Push(&testEvent{}, serviceTraits("websocket").With("channel", "bar")))

		app.Websocket.Remove(a)

		require.Len(t, app.Events.GetEvents(serviceTraits("websocket")), 1)
	})
}
