package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/common"
	"mokapi/kafka"
)

type KafkaClusters map[string]*kafka.Cluster

func (kc KafkaClusters) UpdateConfig(file *common.File) {
	config, ok := file.Data.(*asyncApi.Config)
	if !ok {
		return
	}

	if c, ok := kc[config.Info.Name]; !ok {
		c = kafka.NewCluster(config)
		kc[config.Info.Name] = c
		if err := c.Start(); err != nil {
			log.Errorf("unable to start kafka cluster %v: %v", config.Info.Name, err)
		}
	} else {
		c.Update(config)
	}
}

func (kc KafkaClusters) Stop() {
	for _, c := range kc {
		c.Close()
	}
}

func (s *Server) updateAsyncConfig(config *asyncApi.Config) {

}
