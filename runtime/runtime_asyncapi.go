package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"
)

// mergeAsyncApiConfigs patches the AsyncAPI files of a Kafka cluster, MQTT cluster or
// WebSocket service, ordered by file name, into a new config. The new config owns all
// its data, so merging does not change the files and removing a file leaves no trace of it.
func mergeAsyncApiConfigs(configs map[string]*dynamic.Config, reader dynamic.Reader) *asyncapi3.Config {
	var keys []string
	for k := range configs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		x := keys[i]
		y := keys[j]
		return filepath.Base(x) < filepath.Base(y)
	})

	cfg := &asyncapi3.Config{}
	for i, k := range keys {
		p, ok := IsAsyncApiConfig(configs[k])
		if !ok {
			continue
		}
		if i > 0 {
			log.Infof("applying patch for %s: %s", cfg.Info.Name, k)
		}
		cfg.Patch(p)
	}

	if len(configs) > 1 {
		err := cfg.Parse(&dynamic.Config{Data: cfg}, reader)
		if err != nil {
			log.Errorf("failed to parse config: %s", err)
		}
	}
	return cfg
}
