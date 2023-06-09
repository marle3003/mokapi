package runtime

import (
	log "github.com/sirupsen/logrus"
	cfg "mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/swagger"
	"mokapi/engine/common"
	"mokapi/runtime/monitor"
	"net/http"
)

type HttpInfo struct {
	*openapi.Config
	Name    string
	configs map[string]*openapi.Config
}

type HttpConfig struct {
	Name    string
	configs map[string]*openapi.Config
}

type httpHandler struct {
	http *monitor.Http
	next http.Handler
}

func NewHttpInfo(c *cfg.Config) *HttpInfo {
	hc := &HttpInfo{
		configs: map[string]*openapi.Config{},
	}
	hc.AddConfig(c)
	return hc
}

func (c *HttpInfo) AddConfig(config *cfg.Config) {
	var oc *openapi.Config
	if sw, ok := config.Data.(*swagger.Config); ok {
		var err error
		oc, err = swagger.Convert(sw)
		if err != nil {
			log.Errorf("unable to convert swagger config to openapi: %v", err)
			return
		}
	} else {
		oc = config.Data.(*openapi.Config)
	}
	if len(c.Name) == 0 {
		c.Name = oc.Info.Name
	}
	key := config.Info.Url.String()
	c.configs[key] = oc
	c.update()
}

func (c *HttpInfo) Handler(http *monitor.Http, emitter common.EventEmitter) http.Handler {
	cfg := c.Config
	h := openapi.NewHandler(cfg, emitter)
	return &httpHandler{http: http, next: h}
}

func (c *HttpInfo) update() {
	cfg := &openapi.Config{}
	cfg.Info.Name = c.Name
	for _, p := range c.configs {
		cfg.Patch(p)
	}

	if len(cfg.Servers) == 0 {
		cfg.Servers = append(cfg.Servers, &openapi.Server{Url: "/"})
	}

	c.Config = cfg
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewHttpContext(r.Context(), h.http)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}

func IsHttpConfig(c *cfg.Config) bool {
	switch c.Data.(type) {
	case *openapi.Config:
		return true
	case *swagger.Config:
		return true
	default:
		return false
	}
}
