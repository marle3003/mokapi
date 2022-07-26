package cmd

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/runtime/events"
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
	events.SetStore(20, events.NewTraits().WithNamespace("http"))
	events.SetStore(20, events.NewTraits().WithNamespace("kafka"))

	app := runtime.New()

	watcher := dynamic.NewConfigWatcher(cfg)

	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	kafka := make(server.KafkaClusters)
	http := make(server.HttpServers)
	mail := make(server.SmtpServers)
	directories := make(server.LdapDirectories)
	scriptEngine := engine.New(watcher, app)

	managerHttp := server.NewHttpManager(http, scriptEngine, certStore, app)
	mangerKafka := server.NewKafkaManager(kafka, scriptEngine, app)
	managerLdap := server.NewLdapDirectoryManager(directories, scriptEngine, certStore, app)

	watcher.AddListener(func(cfg *common.Config) {
		mangerKafka.UpdateConfig(cfg)
		managerHttp.Update(cfg)
		mail.UpdateConfig(cfg, certStore, scriptEngine)
		managerLdap.UpdateConfig(cfg)
		if err := scriptEngine.AddScript(cfg); err != nil {
			panic(err)
		}
	})

	if u, err := api.BuildUrl(cfg.Api); err == nil {
		err = managerHttp.AddService("api", u, api.New(app, cfg.Api), true)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	pool := safe.NewPool(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	s := server.NewServer(pool, app, watcher, kafka, http, mail, directories, scriptEngine)
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
