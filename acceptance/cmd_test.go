package acceptance

import (
	"context"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/feature"
	"mokapi/health"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/directory"
	mail2 "mokapi/providers/mail"
	"mokapi/providers/openapi"
	"mokapi/providers/swagger"
	"mokapi/runtime"
	"mokapi/safe"
	"mokapi/schema/json/generator"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/version"

	log "github.com/sirupsen/logrus"
)

type Cmd struct {
	App *runtime.App

	server *server.Server
}

func Start(cfg *static.Config) (*Cmd, error) {
	log.SetLevel(log.DebugLevel)

	feature.Enable(cfg.Features)

	registerDynamicTypes()

	watcher := server.NewConfigWatcher(cfg)
	app := runtime.New(cfg, watcher)
	generator.SetConfig(cfg.DataGen)

	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	scriptEngine := engine.New(watcher, app, cfg, true)

	http := server.NewHttpManager(scriptEngine, certStore, app)
	kafka := server.NewKafkaManager(scriptEngine, app)
	mailManager := server.NewMailManager(app, scriptEngine, certStore)
	ldap := server.NewLdapDirectoryManager(scriptEngine, certStore, app)

	watcher.AddListener(func(e dynamic.ConfigEvent) {
		kafka.UpdateConfig(e)
		http.Update(e)
		mailManager.UpdateConfig(e)
		ldap.UpdateConfig(e)
		if err := scriptEngine.AddScript(e); err != nil {
			panic(err)
		}
		app.UpdateConfig(e)
	})

	apiHandler := api.New(app, cfg.Api)
	if u, err := api.BuildUrl(cfg.Api); err == nil {
		err = http.AddInternalService("api", u, apiHandler)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	if cfg.Health.Enabled {
		if u, err := health.BuildUrl(cfg.Health); err == nil {
			if cfg.Api.Port == cfg.Health.Port {
				apiHandler.RegisterHealthHandler(u.Path, health.New(cfg.Health))
			} else {
				err = http.AddInternalService("health", u, health.New(cfg.Health))
				if err != nil {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}

	pool := safe.NewPool(context.Background())
	s := server.NewServer(pool, app, watcher, kafka, http, mailManager, ldap, scriptEngine)
	go func() {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	}()

	return &Cmd{
		App:    app,
		server: s,
	}, nil
}

func (cmd *Cmd) Stop() {
	cmd.server.Close()
}

func registerDynamicTypes() {
	dynamic.Register("openapi", func(v version.Version) bool {
		return true
	}, &openapi.Config{})
	dynamic.Register("asyncapi", func(v version.Version) bool {
		return v.Major == 2
	}, &asyncApi.Config{})
	dynamic.Register("asyncapi", func(v version.Version) bool {
		return v.Major == 3
	}, &asyncapi3.Config{})
	dynamic.Register("swagger", func(v version.Version) bool {
		return true
	}, &swagger.Config{})
	dynamic.Register("ldap", func(v version.Version) bool {
		return true
	}, &directory.Config{})
	dynamic.Register("smtp", func(v version.Version) bool {
		return true
	}, &mail.Config{})
	dynamic.Register("mail", func(v version.Version) bool {
		return true
	}, &mail2.Config{})
}
