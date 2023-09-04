package cmd

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"
	"mokapi/server"
	"mokapi/server/cert"
)

type Cmd struct {
	App *runtime.App

	server *server.Server
	cancel context.CancelFunc
}

func Start(cfg *static.Config) (*Cmd, error) {
	log.SetLevel(log.DebugLevel)

	if len(cfg.Services) > 0 {
		return nil, errors.New("static configuration Services are no longer supported. Use patching instead.")
	}

	app := runtime.New()

	watcher := dynamic.NewConfigWatcher(cfg)

	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	scriptEngine := engine.New(watcher, app, cfg.Js)

	http := server.NewHttpManager(scriptEngine, certStore, app)
	kafka := server.NewKafkaManager(scriptEngine, app)
	smtp := server.NewSmtpManager(app, scriptEngine, certStore)
	ldap := server.NewLdapDirectoryManager(scriptEngine, certStore, app)

	watcher.AddListener(func(cfg *common.Config) {
		kafka.UpdateConfig(cfg)
		http.Update(cfg)
		smtp.UpdateConfig(cfg)
		ldap.UpdateConfig(cfg)
		if err := scriptEngine.AddScript(cfg); err != nil {
			panic(err)
		}
	})

	if u, err := api.BuildUrl(cfg.Api); err == nil {
		err = http.AddService("api", u, api.New(app, cfg.Api), true)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	pool := safe.NewPool(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	s := server.NewServer(pool, app, watcher, kafka, http, smtp, ldap, scriptEngine)
	s.StartAsync(ctx)

	return &Cmd{
		App:    app,
		server: s,
		cancel: cancel,
	}, nil
}

func (cmd *Cmd) Stop() {
	cmd.cancel()
	cmd.server.Close()
}
