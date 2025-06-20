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

type Broker interface {
	Addr() string
	Start()
	Stop()
}

type kafkaCluster struct {
	brokers map[string]Broker
	cfg     *runtime.KafkaInfo
}

type KafkaManager struct {
	clusters map[string]*kafkaCluster
	emitter  common.EventEmitter
	app      *runtime.App
	m        sync.Mutex
}

func NewKafkaManager(emitter common.EventEmitter, app *runtime.App) *KafkaManager {
	return &KafkaManager{
		clusters: map[string]*kafkaCluster{},
		emitter:  emitter,
		app:      app,
	}
}

func (m *KafkaManager) UpdateConfig(e dynamic.ConfigEvent) {
	// todo: should be IsAsyncConfig and HasKafkaBrokers
	cfg, ok := runtime.IsKafkaConfig(e.Config)
	var info *runtime.KafkaInfo
	if cfg != nil {
		info = m.app.Kafka.Get(cfg.Info.Name)
	}

	if !ok && info == nil {
		return
	}

	if e.Event == dynamic.Delete {
		m.app.Kafka.Remove(e.Config)
		if info.Config == nil {
			m.removeCluster(cfg.Info.Name)
			return
		}
	} else if info == nil {
		var err error
		info, err = m.app.Kafka.Add(e.Config, m.emitter)
		if err != nil {
			log.Errorf("add kafka config %v failed: %v", e.Config.Info.Url, err)
			return
		}
	} else {
		info.AddConfig(e.Config)
	}

	m.addOrUpdateCluster(info)
	log.Debugf("processed %v", e.Config.Info.Path())
}

func (m *KafkaManager) addOrUpdateCluster(cfg *runtime.KafkaInfo) {
	c := m.getOrCreateCluster(cfg)
	c.update(cfg, m.app.Monitor.Kafka)
}

func (m *KafkaManager) getOrCreateCluster(cfg *runtime.KafkaInfo) *kafkaCluster {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.clusters[cfg.Info.Name]
	if !ok {
		log.Infof("adding new kafka cluster '%v'", cfg.Info.Name)
		c = &kafkaCluster{cfg: cfg, brokers: make(map[string]Broker)}
		m.clusters[cfg.Info.Name] = c
	}
	return c
}

func (m *KafkaManager) removeCluster(name string) {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.clusters[name]
	if !ok {
		return
	}
	log.Infof("removing kafka cluster '%v'", name)
	c.close()
	delete(m.clusters, name)
}

func (c *kafkaCluster) update(cfg *runtime.KafkaInfo, kafkaMonitor *monitor.Kafka) {
	c.updateBrokers(cfg, kafkaMonitor)
}

func (c *kafkaCluster) updateBrokers(cfg *runtime.KafkaInfo, kafkaMonitor *monitor.Kafka) {
	brokers := c.brokers
	c.brokers = make(map[string]Broker)
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
			log.Infof("adding new kafka broker '%v' on port %v to cluster '%v'", name, port, cfg.Info.Name)
			broker = service.NewKafkaBroker(port, cfg.Handler(kafkaMonitor))
			broker.Start()
		}
		c.brokers[port] = broker
	}

	for name, broker := range brokers {
		log.Infof("removing kafka broker '%v' on port %v from cluster '%v'", name, broker.Addr(), cfg.Info.Name)
		broker.Stop()
	}
}

func (c *kafkaCluster) close() {
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
