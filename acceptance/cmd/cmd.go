package cmd

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/directory"
	"mokapi/config/dynamic/mail"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/swagger"
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

	registerDynamicTypes()
	app := runtime.New()

	watcher := server.NewConfigWatcher(cfg)

	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	scriptEngine := engine.New(watcher, app, cfg.Js)

	http := server.NewHttpManager(scriptEngine, certStore, app)
	kafka := server.NewKafkaManager(scriptEngine, app)
	smtp := server.NewSmtpManager(app, scriptEngine, certStore)
	ldap := server.NewLdapDirectoryManager(scriptEngine, certStore, app)

	watcher.AddListener(func(cfg *dynamic.Config) {
		kafka.UpdateConfig(cfg)
		http.Update(cfg)
		smtp.UpdateConfig(cfg)
		ldap.UpdateConfig(cfg)
		if err := scriptEngine.AddScript(cfg); err != nil {
			panic(err)
		}
		app.AddConfig(cfg)
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

func registerDynamicTypes() {
	dynamic.Register("openapi", &openapi.Config{})
	dynamic.Register("asyncapi", &asyncApi.Config{})
	dynamic.Register("swagger", &swagger.Config{})
	dynamic.Register("ldap", &directory.Config{})
	dynamic.Register("smtp", &mail.Config{})
}
