package server_test

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"
	"mokapi/server"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	pool := safe.NewPool(context.Background())
	cfg := &static.Config{}

	app := runtime.New(cfg)
	watcher := server.NewConfigWatcher(cfg)
	kafka := &server.KafkaManager{}
	http := &server.HttpManager{}
	mail := &server.MailManager{}
	ldap := &server.LdapDirectoryManager{}
	e := engine.NewEngine()

	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	s := server.NewServer(pool, app, watcher, kafka, http, mail, ldap, e)
	go func() {
		err := s.Start()
		require.NoError(t, err)
	}()

	time.Sleep(time.Second)

	s.Close()

	time.Sleep(time.Second)

	require.Len(t, hook.Entries, 1)
}
