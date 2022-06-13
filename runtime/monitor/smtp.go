package monitor

import (
	"mokapi/runtime/metrics"
)

type Smtp struct {
	Mails *metrics.CounterMap
}
