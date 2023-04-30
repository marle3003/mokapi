package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	config "mokapi/config/dynamic/common"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/server/service"
	"net/url"
)

type clusters map[string]*cluster

type cluster struct {
	brokers map[string]*service.KafkaBroker
	store   *store.Store
}

type KafkaManager struct {
	clusters clusters
	emitter  common.EventEmitter
	app      *runtime.App
}

func NewKafkaManager(emitter common.EventEmitter, app *runtime.App) *KafkaManager {
	return &KafkaManager{
		clusters: clusters{},
		emitter:  emitter,
		app:      app,
	}
}

func (m *KafkaManager) UpdateConfig(c *config.Config) {
	cfg, ok := c.Data.(*asyncApi.Config)
	if !ok {
		return
	}

	m.addOrUpdateCluster(cfg)
	log.Debugf("processed %v", c.Url.String())
}

func (m *KafkaManager) addOrUpdateCluster(cfg *asyncApi.Config) {
	c := m.getOrCreateCluster(cfg)
	c.update(cfg, m.app.Monitor.Kafka)
}

func (m *KafkaManager) getOrCreateCluster(cfg *asyncApi.Config) *cluster {
	c, ok := m.clusters[cfg.Info.Name]
	if !ok {
		log.Infof("adding new kafka cluster '%v'", cfg.Info.Name)
		c = &cluster{store: store.NewEmpty(m.emitter), brokers: make(map[string]*service.KafkaBroker)}
		m.clusters[cfg.Info.Name] = c
		m.app.AddKafka(cfg, c.store)
	}
	return c
}

func (c *cluster) update(cfg *asyncApi.Config, kafkaMonitor *monitor.Kafka) {
	c.store.Update(cfg)
	c.updateBrokers(cfg, kafkaMonitor)
}

func (c *cluster) updateBrokers(cfg *asyncApi.Config, kafkaMonitor *monitor.Kafka) {
	handler := runtime.NewKafkaHandler(kafkaMonitor, c.store)
	brokers := c.brokers
	c.brokers = make(map[string]*service.KafkaBroker)
	for name, server := range cfg.Servers {
		port, err := getPortFromUrl(server.Url)
		if err != nil {
			log.Errorf("unable to start broker %v for cluster %v: ", server.Url, cfg.Info.Name)
			continue
		}

		broker, found := brokers[port]
		if found {
			delete(brokers, port)
		} else {
			log.Infof("adding new kafka broker '%v' on port %v", name, port)
			broker = service.NewKafkaBroker(port, handler)
			broker.Start()
		}
		c.brokers[port] = broker
	}

	for name, broker := range brokers {
		log.Infof("removing kafka broker '%v'", name)
		broker.Stop()
	}
}

func (c *cluster) close() {
	for _, b := range c.brokers {
		b.Stop()
	}
}

func (m *KafkaManager) Stop() {
	for _, c := range m.clusters {
		c.close()
	}
}

func getPortFromUrl(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil || len(u.Host) == 0 {
		u, err = url.Parse("//" + urlString)
		if err != nil {
			return "", err
		}
	}

	port := u.Port()
	if len(port) == 0 {
		port = "9092"
	}

	return port, nil
}
