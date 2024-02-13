package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/kafka"
	"mokapi/runtime/monitor"
	"path/filepath"
	"sort"
)

type KafkaInfo struct {
	*asyncApi.Config
	*store.Store
	configs map[string]*dynamic.Config
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func NewKafkaInfo(c *dynamic.Config, store *store.Store) *KafkaInfo {
	hc := &KafkaInfo{
		configs: map[string]*dynamic.Config{},
		Store:   store,
	}
	hc.AddConfig(c)
	return hc
}

func (c *KafkaInfo) AddConfig(config *dynamic.Config) {
	key := config.Info.Url.String()
	c.configs[key] = config
	c.update()
}

func (c *KafkaInfo) update() {
	var keys []string
	for k := range c.configs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		x := keys[i]
		y := keys[j]
		return filepath.Base(x) < filepath.Base(y)
	})

	cfg := getKafkaConfig(c.configs[keys[0]])
	for _, k := range keys[1:] {
		p := getKafkaConfig(c.configs[k])
		cfg.Patch(p)
	}

	c.Config = cfg
	c.Store.Update(cfg)
}

func (c *KafkaInfo) Handler(kafka *monitor.Kafka) kafka.Handler {
	return &KafkaHandler{kafka: kafka, next: c.Store}
}

func (c *KafkaInfo) Configs() []*dynamic.Config {
	var r []*dynamic.Config
	for _, config := range c.configs {
		r = append(r, config)
		//r = append(r, config.Refs.List()...)
	}
	return r
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func IsKafkaConfig(c *dynamic.Config) bool {
	_, ok := c.Data.(*asyncApi.Config)
	return ok
}

func getKafkaConfig(c *dynamic.Config) *asyncApi.Config {
	return c.Data.(*asyncApi.Config)
}
