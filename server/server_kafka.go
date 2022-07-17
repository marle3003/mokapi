package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	config "mokapi/config/dynamic/common"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/service"
	"net"
	"net/url"
)

type KafkaClusters map[string]*Cluster

type Cluster struct {
	Name  string
	Store *store.Store
	close map[string]func()
}

type KafkaManager struct {
	Clusters KafkaClusters
	brokers  map[string]*service.KafkaBroker //map[port]

	emitter common.EventEmitter
	app     *runtime.App
}

func NewKafkaManager(clusters KafkaClusters, emitter common.EventEmitter, app *runtime.App) *KafkaManager {
	return &KafkaManager{
		Clusters: clusters,
		emitter:  emitter,
		app:      app,
		brokers:  make(map[string]*service.KafkaBroker),
	}
}

func (m KafkaManager) UpdateConfig(c *config.Config) {
	config, ok := c.Data.(*asyncApi.Config)
	if !ok {
		return
	}

	cluster, ok := m.Clusters[config.Info.Name]
	if !ok {
		log.Infof("adding new kafka cluster '%v'", config.Info.Name)
		cluster = &Cluster{
			Name:  config.Info.Name,
			Store: store.New(config),
			close: make(map[string]func()),
		}
		m.Clusters[config.Info.Name] = cluster
		m.app.AddKafka(config, cluster.Store)
	} else {
		cluster.Store.Update(config)
	}

	for _, s := range config.Servers {
		m.AddOrUpdateBroker(s.Url, cluster)
	}

skip:
	for u, f := range cluster.close {
		for _, s := range config.Servers {
			if u == s.Url {
				continue skip
			}
		}
		f()
	}

	log.Debugf("processed %v", c.Url.String())
}

func (kc KafkaClusters) Stop() {
	if len(kc) > 0 {
		log.Debug("stopping kafka clusters")
	}
	for _, c := range kc {
		c.Close()
	}
}

func (m *KafkaManager) AddOrUpdateBroker(url string, cluster *Cluster) {
	host, port, err := parseKafkaUrl(url)
	if err != nil {
		log.Errorf("error %v: %v", url, err.Error())
		return
	}

	addr := fmt.Sprintf("%v:%v", host, port)
	b, ok := m.brokers[addr]
	if !ok {
		b = service.NewKafkaBroker(port)
		b.Start()
		m.brokers[addr] = b
	}
	b.Add(addr, runtime.NewKafkaMonitor(m.app.Monitor.Kafka, cluster.Store))
	cluster.close[url] = func() { b.Remove(addr) }
}

func (c *Cluster) Close() {
	for _, f := range c.close {
		f()
	}
	c.Store.Close()
}

func parseKafkaUrl(s string) (host, port string, err error) {
	u, err := url.Parse(s)
	if err != nil {
		host, port, err = net.SplitHostPort(s)
	} else {
		host = u.Hostname()
		port = u.Port()
	}

	if len(port) == 0 {
		port = "9092"
	}

	return
}
