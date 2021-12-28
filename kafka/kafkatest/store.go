package kafkatest

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/kafka/store"
)

type StoreConfig struct {
	Brokers []string
	Topics  []TopicConfig
}

type TopicConfig struct {
	Name       string
	Partitions int
}

func NewStore(c StoreConfig) *store.Store {
	opt := make([]asyncapitest.ConfigOptions, 0)
	for _, t := range c.Topics {
		opt = append(opt, asyncapitest.WithChannel(t.Name,
			asyncapitest.WithChannelBinding("partitions", fmt.Sprintf("%v", t.Partitions))))
	}
	for _, b := range c.Brokers {
		opt = append(opt, asyncapitest.WithServer(b, "kafka", b))
	}

	return store.New(asyncapitest.NewConfig(opt...))
}
