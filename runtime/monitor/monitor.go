package monitor

import "mokapi/runtime/metrics"

type Monitor struct {
	Http *Http
}

func New() *Monitor {
	return &Monitor{
		Http: &Http{
			HttpMetrics: metrics.NewHttp(),
		},
	}
}
