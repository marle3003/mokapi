package server

// type Manager struct {
// 	entryPoints map[string]*EntryPoint
// }

// func NewManager() *Manager {
// 	return &Manager{}
// }

// func (m *Manager) Build(cfg *static.Config) map[string]*EntryPoint {
// 	m.entryPoints = make(map[string]*EntryPoint)
// 	//m.buildServices(cfg.Services)
// 	return m.entryPoints
// }

// func (m *Manager) buildServices(services map[string]*static.Service) {
// 	for name, service := range services {
// 		log.WithFields(log.Fields{"service": name}).Info("Building service")

// 		api := &dynamic.Api{}
// 		error := service.ApiProviders.File.Provide(api)
// 		if error != nil {
// 			log.WithFields(log.Fields{"service": service, "error": error}).Error("error in provider")
// 			continue
// 		}

// 		dataProvider, error := m.getDataProvider(service)
// 		if error != nil {
// 			log.WithFields(log.Fields{"service": service, "error": error}).Error("error in provider")
// 			continue
// 		}

// 		apiManager := NewApiManager(api, dataProvider)
// 		handler := apiManager.Build()

// 		for _, server := range api.Servers {
// 			entryPoint, error := m.GetEntryPoint(server)
// 			if error == nil {
// 				entryPoint.handler.AddHandler(server.GetPath(), handler)
// 			}
// 		}
// 	}
// }

// func (m *Manager) getDataProvider(service *static.Service) (data.DataProvider, error) {
// 	if service.DataProviders != nil {
// 		if service.DataProviders.File != nil {
// 			store := make(map[interface{}]interface{})
// 			error := service.DataProviders.File.Provide(store)
// 			if error != nil {
// 				return nil, error
// 			}
// 			return data.NewStaticDataProvider(store), nil
// 		}
// 	}

// 	return data.NewRandomDataProvider(), nil
// }

// func (m *Manager) GetEntryPoint(server *dynamic.Server) (*EntryPoint, error) {
// 	name := fmt.Sprintf("%s:%v", server.GetHost(), server.GetPort())
// 	if e, ok := m.entryPoints[name]; ok {
// 		return e, nil
// 	}

// 	entryPoint, error := NewEntryPoint(server)
// 	if error != nil {
// 		return nil, error
// 	}

// 	m.entryPoints[name] = entryPoint
// 	return entryPoint, nil
// }

// type EntryPoint struct {
// 	host    string
// 	port    int
// 	handler *handlers.EntryPointHandler
// }

// func NewEntryPoint(server *dynamic.Server) (*EntryPoint, error) {
// 	entryPoint := &EntryPoint{host: server.GetHost(), port: server.GetPort()}

// 	if len(entryPoint.host) == 0 || entryPoint.port == -1 {
// 		return nil, fmt.Errorf("Invalid entrypoint host %s, port %v", entryPoint.host, entryPoint.port)
// 	}

// 	entryPoint.handler = handlers.NewEntryPointHandler()

// 	return entryPoint, nil
// }
