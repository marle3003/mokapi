package server

import (
	"mokapi/config/dynamic"
)

type ConfigWatcher struct {
	provider      dynamic.Provider
	configuration *dynamic.Configuration

	channel     chan dynamic.ConfigMessage
	stopChannel chan bool

	listeners []func(config *dynamic.Configuration)
}

func NewConfigWatcher(provider dynamic.Provider) *ConfigWatcher {
	return &ConfigWatcher{
		provider:      provider,
		channel:       make(chan dynamic.ConfigMessage, 100),
		configuration: dynamic.NewConfiguration(),
		stopChannel:   make(chan bool)}
}

func (w *ConfigWatcher) AddListener(listener func(config *dynamic.Configuration)) {
	if w.listeners == nil {
		w.listeners = make([]func(*dynamic.Configuration), 0)
	}
	w.listeners = append(w.listeners, listener)
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
				if !ok || isEmpty(configMessage.Config) {
					break
				}

				if configMessage.Config.OpenApi != nil {
					w.configuration.OpenApi[configMessage.Key] = configMessage.Config.OpenApi
				}
				if configMessage.Config.Ldap != nil {
					w.configuration.Ldap[configMessage.Key] = configMessage.Config.Ldap
				}
				if configMessage.Config.AsyncApi != nil {
					w.configuration.AsyncApi[configMessage.Key] = configMessage.Config.AsyncApi
				}

				for _, listener := range w.listeners {
					listener(w.configuration)
				}
			}
		}
	}()
}

func isEmpty(config *dynamic.ConfigurationItem) bool {
	return config.OpenApi == nil && config.Ldap == nil && config.AsyncApi == nil
}
