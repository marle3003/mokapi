package models

type Application struct {
	Services map[string]*ServiceInfo
}

type ServiceInfo struct {
	Service *Service
	Status  string
}

func (a Application) AddOrUpdateService(s *Service) {
	if si, ok := a.Services[s.Name]; ok {
		si.Service = s
	} else {
		a.Services[s.Name] = NewServiceInfo(s)
	}
}

func NewApplication() *Application {
	return &Application{Services: make(map[string]*ServiceInfo)}
}

func NewServiceInfo(s *Service) *ServiceInfo {
	return &ServiceInfo{Service: s}
}
