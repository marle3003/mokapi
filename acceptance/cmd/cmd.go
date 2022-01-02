package cmd

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/safe"
	"mokapi/server"
	"mokapi/server/cert"
)

type Cmd struct {
	server *server.Server
	cancel context.CancelFunc
}

func Start(cfg *static.Config) (*Cmd, error) {
	log.SetLevel(log.DebugLevel)
	watcher := dynamic.NewConfigWatcher(cfg)

	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	kafka := make(server.KafkaClusters)
	web := make(server.WebBindings)
	e := engine.New(watcher)
	watcher.AddListener(func(c *common.File) {
		kafka.UpdateConfig(c)
	})
	watcher.AddListener(func(c *common.File) {
		web.UpdateConfig(c, certStore, e)
	})
	watcher.AddListener(func(f *common.File) {
		if s, ok := f.Data.(*script.Script); ok {
			err := e.AddScript(f.Url.String(), s.Code)
			if err != nil {
				log.Error(err)
			}
		}
	})

	pool := safe.NewPool(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	s := server.NewServer(pool, watcher, kafka, web, e)
	s.StartAsync(ctx)

	return &Cmd{
		server: s,
		cancel: cancel,
	}, nil
}

func (cmd *Cmd) Stop() {
	cmd.cancel()
	cmd.server.Close()
}
