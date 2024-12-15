package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/server/service"
	"net/url"
	"sync"
)

type clusters map[string]*cluster

type cluster struct {
	brokers map[string]*service.KafkaBroker
	cfg     *runtime.KafkaInfo
}

type KafkaManager struct {
	clusters clusters
	emitter  common.EventEmitter
	app      *runtime.App
	m        sync.Mutex
}

func NewKafkaManager(emitter common.EventEmitter, app *runtime.App) *KafkaManager {
	return &KafkaManager{
		clusters: clusters{},
		emitter:  emitter,
		app:      app,
	}
}

func (m *KafkaManager) UpdateConfig(c *dynamic.Config) {
	if !runtime.IsKafkaConfig(c) {
		return
	}
	cfg, err := m.app.AddKafka(c, m.emitter)
	if err != nil {
		log.Errorf("add kafka config %v failed: %v", c.Info.Url, err)
	}

	m.addOrUpdateCluster(cfg)
	log.Debugf("processed %v", c.Info.Path())
}

func (m *KafkaManager) addOrUpdateCluster(cfg *runtime.KafkaInfo) {
	c := m.getOrCreateCluster(cfg)
	c.update(cfg, m.app.Monitor.Kafka)
}

func (m *KafkaManager) getOrCreateCluster(cfg *runtime.KafkaInfo) *cluster {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.clusters[cfg.Info.Name]
	if !ok {
		log.Infof("adding new kafka cluster '%v'", cfg.Info.Name)
		c = &cluster{cfg: cfg, brokers: make(map[string]*service.KafkaBroker)}
		m.clusters[cfg.Info.Name] = c
	}
	return c
}

func (c *cluster) update(cfg *runtime.KafkaInfo, kafkaMonitor *monitor.Kafka) {
	c.updateBrokers(cfg, kafkaMonitor)
}

func (c *cluster) updateBrokers(cfg *runtime.KafkaInfo, kafkaMonitor *monitor.Kafka) {
	brokers := c.brokers
	c.brokers = make(map[string]*service.KafkaBroker)
	for name, server := range cfg.Servers {
		if server == nil || server.Value == nil {
			continue
		}
		port, err := getPortFromUrl(server.Value.Host)
		if err != nil {
			log.Errorf("unable to start broker %v for cluster %v: ", server.Value.Host, cfg.Info.Name)
			continue
		}

		broker, found := brokers[port]
		if found {
			delete(brokers, port)
		} else {
			log.Infof("adding new kafka broker '%v' on port %v", name, port)
			broker = service.NewKafkaBroker(port, cfg.Handler(kafkaMonitor))
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
