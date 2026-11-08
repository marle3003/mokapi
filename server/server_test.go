package server_test

import (
	"context"
	"io"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"
	"mokapi/server"
	"mokapi/try"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	log.SetOutput(io.Discard)

	pool := safe.NewPool(context.Background())
	cfg := &static.Config{}

	app := runtime.New(cfg, &dynamictest.Reader{})
	watcher := server.NewConfigWatcher(cfg)
	kafka := &server.KafkaManager{}
	http := &server.HttpManager{}
	mail := &server.MailManager{}
	ldap := &server.LdapDirectoryManager{}
	mqtt := &server.MqttManager{}
	websocket := &server.WebsocketManager{}
	e := engine.NewEngine()

	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	s := server.NewServer(pool, app, watcher, kafka, http, mail, ldap, mqtt, websocket, e)
	stopped := make(chan bool)
	go func() {
		err := s.Start()
		require.NoError(t, err)
		close(stopped)
	}()

	time.Sleep(time.Second)

	s.Close()

	<-stopped

	require.Len(t, hook.Entries, 1)
	require.Equal(t, "stopping server", hook.Entries[0].Message)
}

func TestServer_ShutdownStopsMqttAndWebsocketManagers(t *testing.T) {
	log.SetOutput(io.Discard)

	cfg := &static.Config{}
	app := runtime.New(cfg, &dynamictest.Reader{})
	mqtt := server.NewMqttManager(nil, app)
	websocket := server.NewWebsocketManager(nil, app)
	mqttPort, websocketPort := try.GetFreePort(), try.GetFreePort()
	mqtt.UpdateConfig(mqttConfigEvent(mqttPort))
	websocket.UpdateConfig(websocketConfigEvent(websocketPort))

	s := server.NewServer(safe.NewPool(context.Background()), app, server.NewConfigWatcher(cfg),
		&server.KafkaManager{}, &server.HttpManager{}, &server.MailManager{}, &server.LdapDirectoryManager{},
		mqtt, websocket, engine.NewEngine())
	stopped := make(chan bool)
	go func() {
		_ = s.Start()
		close(stopped)
	}()

	s.Close()
	<-stopped

	requirePortIsFree(t, mqttPort)
	requirePortIsFree(t, websocketPort)
}
