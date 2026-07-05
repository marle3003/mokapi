package server

import (
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/server/service"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type WebsocketManager struct {
	services map[string]*websocketService
	emitter  common.EventEmitter
	app      *runtime.App
	m        sync.Mutex
}

type websocketService struct {
	servers map[string]*service.WebsocketServer
	cfg     *runtime.WebsocketInfo
}

func NewWebsocketManager(emitter common.EventEmitter, app *runtime.App) *WebsocketManager {
	return &WebsocketManager{
		services: map[string]*websocketService{},
		emitter:  emitter,
		app:      app,
	}
}

func (m *WebsocketManager) UpdateConfig(e dynamic.ConfigEvent) {
	cfg, ok := runtime.HasWebsocketServer(e.Config)
	if !ok {
		if cfg == nil || m.services == nil {
			return
		}
		m.removeService(cfg.Info.Name)
		return
	}

	info := m.app.Websocket.Get(cfg.Info.Name)
	if e.Event == dynamic.Delete && !info.HasWebsocketServer() {
		m.app.Kafka.Remove(e.Config)
		if info.Config == nil {
			m.removeService(cfg.Info.Name)
			return
		}
	}

	var err error
	info, err = m.app.Websocket.Add(e.Config, m.emitter)
	if err != nil {
		log.Errorf("add Websocket config %v failed: %v", e.Config.Info.Url, err)
		return
	}

	m.addOrUpdateService(info)
	log.Debugf("processed %v", e.Config.Info.Path())
}

func (m *WebsocketManager) addOrUpdateService(cfg *runtime.WebsocketInfo) {
	c := m.getOrCreateService(cfg)
	c.update(cfg, m.app.Monitor.Websocket)
}

func (m *WebsocketManager) getOrCreateService(cfg *runtime.WebsocketInfo) *websocketService {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.services[cfg.Info.Name]
	if !ok {
		c = &websocketService{cfg: cfg, servers: make(map[string]*service.WebsocketServer)}
		m.services[cfg.Info.Name] = c
	}
	return c
}

func (m *WebsocketManager) removeService(name string) {
	m.m.Lock()
	defer m.m.Unlock()

	c, ok := m.services[name]
	if !ok {
		return
	}
	log.Infof("removing kafka cluster '%v'", name)
	c.close()
	delete(m.services, name)
}

func (c *websocketService) update(cfg *runtime.WebsocketInfo, monitor *monitor.Websocket) {
	c.updateServers(cfg, monitor)
}

func (c *websocketService) updateServers(cfg *runtime.WebsocketInfo, monitor *monitor.Websocket) {
	servers := c.servers
	c.servers = make(map[string]*service.WebsocketServer)
	for it := cfg.Servers.Iter(); it.Next(); {
		name := it.Key()
		server := it.Value()
		if server == nil || server.Value == nil || server.Value.Protocol != "ws" {
			continue
		}
		port, err := getPortFromUrl(server.Value.Host, "1883")
		if err != nil {
			log.Errorf("unable to start Websocket server %v for service %v: ", server.Value.Host, cfg.Info.Name)
			continue
		}

		broker, found := servers[port]
		if found {
			delete(servers, port)
		} else {
			log.Infof("adding new Websocket server '%v' on port %v to '%v'", name, port, cfg.Info.Name)
			h := cfg.Handler(monitor)
			broker = service.NewWebsocketServer(port, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				req = req.WithContext(websocket.NewServerContext(req.Context(), server.Value))
				h.ServeHTTP(w, req)
			}))
			broker.Start()
		}
		c.servers[port] = broker
	}

	for name, broker := range servers {
		log.Infof("removing Websocket server '%v' on port %v from '%v'", name, broker.Addr(), cfg.Info.Name)
		broker.Stop()
	}
}

func (c *websocketService) close() {
	for _, b := range c.servers {
		b.Stop()
	}
}

func (m *WebsocketManager) Stop() {
	for _, c := range m.services {
		c.close()
	}
}
