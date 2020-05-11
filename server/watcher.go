package server

import (
	"mokapi/config/dynamic"
	"mokapi/service"

	log "github.com/sirupsen/logrus"
)

type RootConfig struct {
	openApis map[string]*dynamic.OpenApi
}

type ConfigWatcher struct {
	provider      dynamic.Provider
	configuration *RootConfig

	channel     chan dynamic.ConfigMessage
	stopChannel chan bool

	listeners     []func(*service.Service)
	ldapListeners []func(key string, config *dynamic.Ldap)
}

func NewConfigWatcher(provider dynamic.Provider) *ConfigWatcher {
	return &ConfigWatcher{
		provider:      provider,
		channel:       make(chan dynamic.ConfigMessage, 100),
		configuration: &RootConfig{openApis: make(map[string]*dynamic.OpenApi)},
		stopChannel:   make(chan bool)}
}

func (w *ConfigWatcher) AddListener(listener func(*service.Service)) {
	if w.listeners == nil {
		w.listeners = make([]func(*service.Service), 0)
	}
	w.listeners = append(w.listeners, listener)
}

func (w *ConfigWatcher) AddLdapListener(listener func(key string, config *dynamic.Ldap)) {
	if w.ldapListeners == nil {
		w.ldapListeners = make([]func(string, *dynamic.Ldap), 0)
	}
	w.ldapListeners = append(w.ldapListeners, listener)
}

func (w *ConfigWatcher) Stop() {
	w.stopChannel <- true
}

func (w *ConfigWatcher) Start() {
	go func() {
		w.provider.ProvideService(w.channel)
	}()

	go func() {
		defer func() {
			w.provider.Close()
		}()

		for {
			select {
			case <-w.stopChannel:
				return
			case configMessage, ok := <-w.channel:
				if !ok {
					break
				}

				log.Infof("Processing configuration %v", configMessage.Key)

				if configMessage.Config.OpenApi != nil {
					part := configMessage.Config.OpenApi
					if part == nil || isEmpty(part) {
						break
					}

					if part.Info.Name == "" {
						part.Info.Name = configMessage.Key
					}

					var config *dynamic.OpenApi
					if s, ok := w.configuration.openApis[part.Info.Name]; ok {
						config = s

					} else {
						config = &dynamic.OpenApi{Parts: make(map[string]*dynamic.OpenApiPart)}
						w.configuration.openApis[part.Info.Name] = config
					}

					config.Parts[configMessage.Key] = part

					service := service.CreateService(config)

					for _, listener := range w.listeners {
						listener(service)
					}
				}

				if configMessage.Config.Ldap != nil {
					for _, listener := range w.ldapListeners {
						listener(configMessage.Key, configMessage.Config.Ldap)
					}
				}
			}
		}
	}()
}

func isEmpty(service *dynamic.OpenApiPart) bool {
	if service == nil {
		return true
	}

	if service.Info.Name == "" {
		return true
	}

	return false
}
