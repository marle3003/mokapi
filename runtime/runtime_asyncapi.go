package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"mokapi/sortedmap"

	log "github.com/sirupsen/logrus"
)

// mergeAsyncApiConfigs patches the AsyncAPI files of a Kafka cluster, MQTT cluster or
// WebSocket service, ordered by file name, into a new config. The new config owns all
// its data, so merging does not change the files and removing a file leaves no trace of it.
func mergeAsyncApiConfigs(files []*dynamic.Config, reader dynamic.Reader) *asyncapi3.Config {
	cfg := &asyncapi3.Config{}
	for i, f := range files {
		p, ok := IsAsyncApiConfig(f)
		if !ok {
			continue
		}
		if i > 0 {
			log.Infof("applying patch for %s: %s", cfg.Info.Name, f.Info.Url)
		}
		cfg.Patch(p)
	}

	if len(files) > 1 {
		err := cfg.Parse(&dynamic.Config{Data: cfg}, reader)
		if err != nil {
			log.Errorf("failed to parse config: %s", err)
		}
	}
	return cfg
}

// defaultAsyncApiServer adds the server "mokapi" to cfg if the spec defines no server for the
// protocol usage, so Mokapi can serve bindings without an explicit server. logWhat is the
// log message prefix of the inferred default.
func defaultAsyncApiServer(cfg *asyncapi3.Config, usage string, logWhat string, srv *asyncapi3.Server) {
	if len(cfg.Analyze().Usage(usage).Servers) > 0 {
		return
	}

	log.Infof("%s for '%s' (inferred from bindings)", logWhat, cfg.Info.Name)
	if cfg.Servers == nil {
		cfg.Servers = &sortedmap.LinkedHashMap[string, *asyncapi3.ServerRef]{}
	}
	cfg.Servers.Set("mokapi", &asyncapi3.ServerRef{Value: srv})
}
