package runtime

import (
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Runtime struct {
	Log     *logs.Logs
	Metrics *metrics.Metrics
}

func New() *Runtime {
	return &Runtime{
		Log:     logs.New(),
		Metrics: metrics.New(),
	}
}
