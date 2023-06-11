package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	cfg "mokapi/config/dynamic/common"
	"mokapi/kafka"
	"mokapi/runtime/monitor"
)

type KafkaInfo struct {
	*asyncApi.Config
	*store.Store

	Name    string
	configs map[string]*asyncApi.Config
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func NewKafkaInfo(c *cfg.Config, store *store.Store) *KafkaInfo {
	hc := &KafkaInfo{
		configs: map[string]*asyncApi.Config{},
		Store:   store,
	}
	hc.AddConfig(c)
	return hc
}

func (c *KafkaInfo) AddConfig(config *cfg.Config) {
	ac := config.Data.(*asyncApi.Config)
	if len(c.Name) == 0 {
		c.Name = ac.Info.Name
	}

	key := config.Info.Url.String()
	c.configs[key] = ac
	c.update()
}

func (c *KafkaInfo) update() {
	cfg := &asyncApi.Config{}
	cfg.Info.Name = c.Name
	for _, p := range c.configs {
		cfg.Patch(p)
	}

	c.Config = cfg
	c.Store.Update(cfg)
}

func (c *KafkaInfo) Handler(kafka *monitor.Kafka) kafka.Handler {
	return &KafkaHandler{kafka: kafka, next: c.Store}
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func IsKafkaConfig(c *cfg.Config) bool {
	_, ok := c.Data.(*asyncApi.Config)
	return ok
}
