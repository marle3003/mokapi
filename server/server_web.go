package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/server/cert"
	"mokapi/server/web"
	"strings"
)

type WebBindings map[string]*web.Binding

func (wb WebBindings) UpdateConfig(file *common.File, certStore *cert.Store) {
	config, ok := file.Data.(*openapi.Config)
	if !ok {
		return
	}

	if err := config.Validate(); err != nil {
		log.Warnf("validation error %v: %v", file.Url, err)
		return
	}

	for _, server := range config.Servers {
		_, port, _, err := web.ParseAddress(server.Url)
		if err != nil {
			log.Errorf("%v: %v", err, file.Url)
			continue
		}
		address := fmt.Sprintf(":%v", port)
		binding, found := wb[address]
		if !found {
			if strings.HasPrefix(strings.ToLower(server.Url), "https://") {
				binding = web.NewBindingWithTls(address, certStore)
			} else {
				binding = web.NewBinding(address)
			}
			wb[address] = binding
			binding.Start()
		}
		err = binding.Apply(config)
		if err != nil {
			log.Errorf("error on updating %v: %v", file.Url.String(), err.Error())
			return
		}
	}
	log.Infof("processed config %v", file.Url.String())
}

func (wb WebBindings) Stop() {
	for _, b := range wb {
		b.Stop()
	}
}
