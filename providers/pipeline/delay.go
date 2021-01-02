package pipeline

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"strings"
	"time"
)

type DelayStep struct {
}

type DelayStepExecution struct {
	Type   string
	Median float64
	Sigma  float64
	Lower  int
	Upper  int
	Time   int
	Unit   string

	random *rand.Rand `step:-`
}

func (e *DelayStep) Start() StepExecution {
	return &DelayStepExecution{Unit: "seconds"}
}

func (e *DelayStepExecution) Run(_ StepContext) (interface{}, error) {
	unit, err := e.getUnit()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(e.Type) {
	case "lognormal":
		number := math.Round(math.Exp(e.random.NormFloat64()*e.Sigma) * e.Median)
		delay := time.Duration(number) * unit
		log.Infof("Delay request for %v", delay)
		time.Sleep(delay)
	case "uniform":
		number := rand.Intn(e.Upper-e.Lower) + e.Lower
		delay := time.Duration(number) * unit
		log.Infof("Delay request for %v", delay)
		time.Sleep(delay)
	case "":
		delay := time.Duration(e.Time) * unit
		log.Infof("Delay request for %v", delay)
		time.Sleep(delay)
	}

	return nil, nil
}

func (e *DelayStepExecution) getUnit() (time.Duration, error) {
	switch strings.ToLower(e.Unit) {
	case "seconds":
		return time.Second, nil
	case "minutes":
		return time.Minute, nil
	case "milliseconds":
		return time.Millisecond, nil
	case "nanoseconds":
		return time.Nanosecond, nil
	case "microseconds":
		return time.Microsecond, nil
	case "hours":
		return time.Hour, nil
	default:
		return 0, errors.Errorf("unknown unit '%v'", e.Unit)
	}
}
