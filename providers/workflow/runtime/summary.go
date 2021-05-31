package runtime

import (
	"time"
)

type Summary struct {
	Steps    []StepSummary
	Duration time.Duration
}

type StepSummary struct {
	Name     string
	Log      string
	Duration time.Duration
}
