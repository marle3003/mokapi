package cmd

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/server"
	"mokapi/server/cert"
)

type Cmd struct {
	watcher *dynamic.ConfigWatcher
	kafka   server.KafkaClusters
}

func Start(cfg *static.Config) (*Cmd, error) {
	watcher := dynamic.NewConfigWatcher(cfg.Providers)
	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	kafka := make(server.KafkaClusters)
	web := make(server.WebBindings)
	watcher.AddListener(func(c *common.File) {
		kafka.UpdateConfig(c)
	})
	watcher.AddListener(func(c *common.File) {
		web.UpdateConfig(c, certStore)
	})
	err = watcher.Start()
	if err != nil {
		return nil, err
	}

	return &Cmd{
		watcher: watcher,
		kafka:   kafka,
	}, nil
}

func (cmd *Cmd) Stop() {
	cmd.watcher.Close()
	cmd.kafka.Stop()
}
