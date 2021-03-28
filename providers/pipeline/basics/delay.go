package basics

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"mokapi/providers/pipeline/lang/types"
	"strings"
	"time"
)

type DelayStep struct {
	types.AbstractStep
	random *rand.Rand
}

type DelayStepExecution struct {
	Type  string
	Mean  float64
	Sigma float64
	Lower int
	Upper int
	Time  interface{}
	Unit  string

	step *DelayStep
}

func NewDelayStep() *DelayStep {
	return &DelayStep{
		random: rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (e *DelayStep) Start() types.StepExecution {
	return &DelayStepExecution{Unit: "s", step: e}
}

func (e *DelayStepExecution) Run(_ types.StepContext) (interface{}, error) {
	unit, err := e.getUnit()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(e.Type) {
	case "lognormal":
		number := math.Round(e.step.random.NormFloat64()*e.Sigma + e.Mean)
		delay := time.Duration(number) * unit
		log.Infof("Delay request for %v", delay)
		time.Sleep(delay)
	case "uniform":
		number := rand.Intn(e.Upper-e.Lower) + e.Lower
		delay := time.Duration(number) * unit
		log.Infof("Delay request for %v", delay)
		time.Sleep(delay)
	case "":
		var delay time.Duration
		switch t := e.Time.(type) {
		case string:
			d, err := time.ParseDuration(t)
			if err != nil {
				return nil, err
			}
			delay = d
		case float64:
			delay = time.Duration(t) * unit
		}
		log.Infof("Delay request for %v", delay)
		time.Sleep(delay)
	}

	return nil, nil
}

func (e *DelayStepExecution) getUnit() (time.Duration, error) {
	switch strings.ToLower(e.Unit) {
	case "s":
		return time.Second, nil
	case "m":
		return time.Minute, nil
	case "ms":
		return time.Millisecond, nil
	case "ns":
		return time.Nanosecond, nil
	case "us":
		return time.Microsecond, nil
	case "h":
		return time.Hour, nil
	default:
		return 0, errors.Errorf("unknown unit '%v'", e.Unit)
	}
}
