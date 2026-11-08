package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/try"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// requireStopDoesNotRaceWithUpdateConfig runs Stop while another goroutine keeps updating configs, as the
// config watcher does during shutdown. Without the manager lock in Stop, -race reports it.
func requireStopDoesNotRaceWithUpdateConfig(t *testing.T, stop func(), update func(i, port int)) {
	logrus.SetOutput(io.Discard)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			update(i, try.GetFreePort())
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			stop()
		}
	}()
	wg.Wait()
}

func TestLdapManager_StopWhileConfigsAreUpdated(t *testing.T) {
	m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))

	requireStopDoesNotRaceWithUpdateConfig(t, m.Stop, func(i, port int) {
		m.UpdateConfig(dynamic.ConfigEvent{
			Config: newConfig(fmt.Sprintf("host%d.yml", i), &directory.Config{
				Info:    directory.Info{Name: fmt.Sprintf("host%d", i)},
				Address: fmt.Sprintf(":%d", port),
			}),
		})
	})
}

func TestKafkaManager_StopWhileConfigsAreUpdated(t *testing.T) {
	m := server.NewKafkaManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))

	requireStopDoesNotRaceWithUpdateConfig(t, m.Stop, func(_, port int) { m.UpdateConfig(kafkaConfigEvent(port)) })
}

func TestMqttManager_StopWhileConfigsAreUpdated(t *testing.T) {
	m := server.NewMqttManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))

	requireStopDoesNotRaceWithUpdateConfig(t, m.Stop, func(_, port int) { m.UpdateConfig(mqttConfigEvent(port)) })
}

func TestWebsocketManager_StopWhileConfigsAreUpdated(t *testing.T) {
	m := server.NewWebsocketManager(nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))

	requireStopDoesNotRaceWithUpdateConfig(t, m.Stop, func(_, port int) { m.UpdateConfig(websocketConfigEvent(port)) })
}

func TestMailManager_StopWhileConfigsAreUpdated(t *testing.T) {
	m := newMailManager(t)

	requireStopDoesNotRaceWithUpdateConfig(t, m.Stop, func(_, port int) { m.UpdateConfig(mailConfigEvent("smtp", port)) })
}
