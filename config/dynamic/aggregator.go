package dynamic

// func (p *DataProviders) init(context data.Context) {
// 	if p.File != nil {
// 		p.File.Init(context)
// 	}
// }

// func (service *Service) Merge(other *Service) {
// 	if other.Info.Description != "" {
// 		service.Info.Description = other.Info.Description
// 	}
// 	if other.Info.Version != "" {
// 		service.Info.Version = other.Info.Version
// 	}

// 	for key, value := range other.Components.Schemas {
// 		service.Components.Schemas[key] = value
// 	}

// 	for _, server := range other.Servers {
// 		service.Servers = append(service.Servers, server)
// 	}

// 	for path, endpoint := range other.EndPoints {
// 		if _, ok := service.EndPoints[path]; ok {
// 			log.WithFields(log.Fields{"service": service.Info.Name, "path": path}).Error("Can not merge endpoint path")
// 			continue
// 		}
// 		service.EndPoints[path] = endpoint
// 	}
// }
