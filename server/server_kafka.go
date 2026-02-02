package server

import (
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/server/service"
	"mokapi/sortedmap"
	"net/url"
	"slices"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Broker interface {
	Addr() string
	Start()
	Stop()
}

type kafkaCluster struct {
	brokers map[string]Broker
	cfg     *runtime.KafkaInfo
	monitor *monitor.Kafka
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
	cfg, ok := runtime.IsAsyncApiConfig(e.Config)
	if !ok {
		return
	}

	info := m.app.Kafka.Get(cfg.Info.Name)
	if e.Event == dynamic.Delete || info != nil && !info.HasKafkaServer() {
		m.app.Kafka.Remove(e.Config)
		if info.Config == nil {
			m.removeCluster(cfg.Info.Name)
			return
		}
	}
	var servers *sortedmap.LinkedHashMap[string, *asyncapi3.ServerRef]
	if info != nil {
		servers = info.Servers
	}
	var err error
	info, err = m.app.Kafka.Add(e.Config, m.emitter)
	if err != nil {
		log.Errorf("add kafka config %v failed: %v", e.Config.Info.Url, err)
		return
	}

	c := m.getOrCreateCluster(info)
	c.updateBrokers(info, servers)

	log.Debugf("processed %v", e.Config.Info.Path())
}

func (m *KafkaManager) getOrCreateCluster(cfg *runtime.KafkaInfo) *kafkaCluster {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.clusters[cfg.Info.Name]
	if !ok {
		log.Infof("adding new kafka cluster '%v'", cfg.Info.Name)
		c = &kafkaCluster{cfg: cfg, brokers: make(map[string]Broker), monitor: m.app.Monitor.Kafka}
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

func (c *kafkaCluster) updateBrokers(cfg *runtime.KafkaInfo, old *sortedmap.LinkedHashMap[string, *asyncapi3.ServerRef]) {
	servers := cfg.Servers.Values()

	for it := old.Iter(); it.Next(); {
		name := it.Key()
		server := it.Value()
		if !slices.ContainsFunc(servers, func(s *asyncapi3.ServerRef) bool {
			return s.Value != nil && server.Value != nil && s.Value.Host == server.Value.Host
		}) {
			port, _ := getPortFromUrl(server.Value.Host)
			if b, ok := c.brokers[port]; ok {
				log.Infof("removing kafka broker '%v' on port %v from cluster '%v'", name, b.Addr(), cfg.Info.Name)
				b.Stop()
				delete(c.brokers, port)
			}
		}
	}

	for it := cfg.Servers.Iter(); it.Next(); {
		name := it.Key()
		server := it.Value()
		if server == nil || server.Value == nil {
			continue
		}
		port, err := getPortFromUrl(server.Value.Host)
		if err != nil {
			log.Errorf("unable to start broker %v for cluster %v: ", server.Value.Host, cfg.Info.Name)
			continue
		}

		broker, found := c.brokers[port]
		if !found {
			log.Infof("adding new kafka broker '%v' on port %v to cluster '%v'", name, port, cfg.Info.Name)
			broker = service.NewKafkaBroker(port, cfg.Handler(c.monitor))
			broker.Start()
		}
		c.brokers[port] = broker
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
