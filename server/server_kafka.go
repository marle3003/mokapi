package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/kafka/schema"
)

func (s *Server) updateAsyncConfig(config *asyncApi.Config) {
	kafkaSchema := toKafkaSchema(config)
	if c, ok := s.kafkaClusters[config.Info.Name]; !ok {
		c = kafka.NewCluster(kafkaSchema)
		s.kafkaClusters[config.Info.Name] = c
		if err := c.Start(); err != nil {
			log.Errorf("unable to start kafka cluster %v: %v", config.Info.Name, err)
		}
	} else {
		c.Update(kafkaSchema)
	}
}

func toKafkaSchema(config *asyncApi.Config) schema.Cluster {
	s := schema.New()
	// todo: order of servers
	for _, server := range config.Servers {
		s.Brokers = append(s.Brokers, schema.Broker{
			Id:   0,
			Host: server.GetHost(),
			Port: server.GetPort(),
		})
	}

	for name := range config.Channels {
		s.Topics = append(s.Topics, schema.Topic{
			Name: name,
			Partitions: []schema.Partition{
				{
					Index:    0,
					Replicas: []int{0},
				},
			},
		})
	}

	return s
}
