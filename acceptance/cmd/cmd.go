package cmd

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/server"
)

type Cmd struct {
	watcher *dynamic.ConfigWatcher
	kafka   server.KafkaClusters
}

func Start(cfg *static.Config) *Cmd {
	watcher := dynamic.NewConfigWatcher(cfg.Providers)
	kafka := make(server.KafkaClusters)
	watcher.AddListener(func(c *common.File) {
		kafka.UpdateConfig(c)
	})
	watcher.Start()

	return &Cmd{
		watcher: watcher,
		kafka:   kafka,
	}
}

func (cmd *Cmd) Stop() {
	cmd.watcher.Close()
	cmd.kafka.Stop()
}
