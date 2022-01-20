package cmd

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/script"
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
	app := runtime.New()

	watcher := dynamic.NewConfigWatcher(cfg)

	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	kafka := make(server.KafkaClusters)
	http := make(server.HttpServers)
	mail := make(server.SmtpServers)
	scriptEngine := engine.New(watcher)

	managerHttp := server.NewHttpManager(http, scriptEngine, certStore, app)

	watcher.AddListener(func(file *common.File) {
		kafka.UpdateConfig(file)
		managerHttp.Update(file)
		mail.UpdateConfig(file, certStore, scriptEngine)
	})
	watcher.AddListener(func(file *common.File) {
		if s, ok := file.Data.(*script.Script); ok {
			err := scriptEngine.AddScript(file.Url.String(), s.Code)
			if err != nil {
				log.Error(err)
			}
		}
	})

	pool := safe.NewPool(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	s := server.NewServer(pool, watcher, kafka, http, mail, scriptEngine)
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
