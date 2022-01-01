package cmd

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/server"
	"mokapi/server/cert"
)

type Cmd struct {
	watcher *dynamic.ConfigWatcher
	kafka   server.KafkaClusters
	pool    *safe.Pool
}

func Start(cfg *static.Config) (*Cmd, error) {
	pool := safe.NewPool(context.Background())
	watcher := dynamic.NewConfigWatcher(cfg)

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
	err = watcher.Start(pool)
	if err != nil {
		return nil, err
	}

	return &Cmd{
		watcher: watcher,
		kafka:   kafka,
		pool:    pool,
	}, nil
}

func (cmd *Cmd) Stop() {
	cmd.pool.Stop()
	cmd.kafka.Stop()
}
