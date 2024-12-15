package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/monitor"
	"path/filepath"
	"sort"
)

type KafkaInfo struct {
	*asyncapi3.Config
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

	var cfg *asyncapi3.Config
	for _, k := range keys {
		p, err := getKafkaConfig(c.configs[k])
		if err != nil {
			log.Errorf("patch %v failed: %v", c.configs[k].Info.Url, err)
		}
		if cfg == nil {
			cfg = p
		} else {
			cfg.Patch(p)
		}
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
	}
	return r
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}

func IsKafkaConfig(c *dynamic.Config) bool {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return true
	}
	if _, ok := c.Data.(*asyncApi.Config); ok {
		return true
	}
	return false
}

func getKafkaConfig(c *dynamic.Config) (*asyncapi3.Config, error) {
	if _, ok := c.Data.(*asyncapi3.Config); ok {
		return c.Data.(*asyncapi3.Config), nil
	} else {
		old := c.Data.(*asyncApi.Config)
		return old.Convert()
	}
}
