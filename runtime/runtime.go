package runtime

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/monitor"
)

type App struct {
	Http map[string]*HttpInfo

	Monitor *monitor.Monitor
}

func New() *App {
	return &App{
		Monitor: monitor.New(),
	}
}

func (r *App) AddHttp(c *openapi.Config) {
	if len(r.Http) == 0 {
		r.Http = make(map[string]*HttpInfo)
	}
	r.Http[c.Info.Name] = &HttpInfo{Config: c}
}
