package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/server/service"
	"sync"
)

type MqttManager struct {
	clusters map[string]*mqttCluster
	emitter  common.EventEmitter
	app      *runtime.App
	m        sync.Mutex
}

type mqttCluster struct {
	brokers map[string]Broker
	cfg     *runtime.MqttInfo
}

func NewMqttManager(emitter common.EventEmitter, app *runtime.App) *MqttManager {
	return &MqttManager{
		clusters: map[string]*mqttCluster{},
		emitter:  emitter,
		app:      app,
	}
}

func (m *MqttManager) UpdateConfig(e dynamic.ConfigEvent) {
	// todo: should be IsAsyncConfig and HasMqttBrokers
	cfg, ok := runtime.IsMqttConfig(e.Config)
	if !ok {
		return
	}

	info := m.app.Mqtt.Get(cfg.Info.Name)
	if e.Event == dynamic.Delete {
		m.app.Kafka.Remove(e.Config)
		if info.Config == nil {
			m.removeCluster(cfg.Info.Name)
			return
		}
	} else if info == nil {
		var err error
		info, err = m.app.Mqtt.Add(e.Config, m.emitter)
		if err != nil {
			log.Errorf("add MQTT config %v failed: %v", e.Config.Info.Url, err)
			return
		}
	} else {
		info.AddConfig(e.Config)
	}

	m.addOrUpdateCluster(info)
	log.Debugf("processed %v", e.Config.Info.Path())
}

func (m *MqttManager) addOrUpdateCluster(cfg *runtime.MqttInfo) {
	c := m.getOrCreateCluster(cfg)
	c.update(cfg, m.app.Monitor.Mqtt)
}

func (m *MqttManager) getOrCreateCluster(cfg *runtime.MqttInfo) *mqttCluster {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.clusters[cfg.Info.Name]
	if !ok {
		c = &mqttCluster{cfg: cfg, brokers: make(map[string]Broker)}
		m.clusters[cfg.Info.Name] = c
	}
	return c
}

func (m *MqttManager) removeCluster(name string) {
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

func (c *mqttCluster) update(cfg *runtime.MqttInfo, monitor *monitor.Mqtt) {
	c.updateBrokers(cfg, monitor)
}

func (c *mqttCluster) updateBrokers(cfg *runtime.MqttInfo, monitor *monitor.Mqtt) {
	brokers := c.brokers
	c.brokers = make(map[string]Broker)
	for name, server := range cfg.Servers {
		if server == nil || server.Value == nil {
			continue
		}
		port, err := getPortFromUrl(server.Value.Host)
		if err != nil {
			log.Errorf("unable to start MQTT broker %v for cluster %v: ", server.Value.Host, cfg.Info.Name)
			continue
		}

		broker, found := brokers[port]
		if found {
			delete(brokers, port)
		} else {
			log.Infof("adding new MQTT broker '%v' on port %v to '%v'", name, port, cfg.Info.Name)
			broker = service.NewMqttBroker(port, cfg.Handler(monitor))
			broker.Start()
		}
		c.brokers[port] = broker
	}

	for name, broker := range brokers {
		log.Infof("removing MQTT broker '%v' on port %v from '%v'", name, broker.Addr(), cfg.Info.Name)
		broker.Stop()
	}
}

func (c *mqttCluster) close() {
	for _, b := range c.brokers {
		b.Stop()
	}
}

func (m *MqttManager) Stop() {
	for _, c := range m.clusters {
		c.close()
	}
}
