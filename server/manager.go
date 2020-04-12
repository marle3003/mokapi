package server

import (
	"fmt"
	"mokapi/config"
	"mokapi/config/static"
	"mokapi/server/handlers"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	entryPoints map[string]*EntryPoint
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Build(cfg *static.Config) map[string]*EntryPoint {
	m.entryPoints = make(map[string]*EntryPoint)
	m.buildServices(cfg.Services)
	return m.entryPoints
}

func (m *Manager) buildServices(services map[string]*static.Service) {
	for _, service := range services {
		serviceHandler := handlers.NewServiceHandler()

		api, error := service.ApiProviders.File.Provide()
		if error != nil {
			log.WithFields(log.Fields{"service": service, "error": error}).Error("error in provider")
			continue
		}

		for path, endpoint := range api.EndPoints {
			endpointHandler := handlers.NewEndpointHandler(endpoint)
			serviceHandler.AddEndpoint(path, endpointHandler)
		}

		for _, server := range api.Servers {
			entryPoint, error := m.GetEntryPoint(server)
			if error == nil {
				entryPoint.handler.AddService(server.GetPath(), serviceHandler)
			}
		}
	}
}

func (m *Manager) GetEntryPoint(server *config.Server) (*EntryPoint, error) {
	name := fmt.Sprintf("%s:%v", server.GetHost(), server.GetPort())
	if e, ok := m.entryPoints[name]; ok {
		return e, nil
	}

	entryPoint, error := NewEntryPoint(server)
	if error != nil {
		return nil, error
	}

	m.entryPoints[name] = entryPoint
	return entryPoint, nil
}

type EntryPoint struct {
	host    string
	port    int
	handler *handlers.EntryPointHandler
}

func NewEntryPoint(server *config.Server) (*EntryPoint, error) {
	entryPoint := &EntryPoint{host: server.GetHost(), port: server.GetPort()}

	if len(entryPoint.host) == 0 || entryPoint.port == -1 {
		return nil, fmt.Errorf("Invalid entrypoint host %s, port %v", entryPoint.host, entryPoint.port)
	}

	entryPoint.handler = handlers.NewEntryPointHandler()

	return entryPoint, nil
}
