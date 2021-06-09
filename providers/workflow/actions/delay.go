package actions

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"mokapi/providers/workflow/runtime"
	"strings"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().Unix()))

type Delay struct {
}

func (e *Delay) Run(ctx *runtime.ActionContext) error {
	t, _ := ctx.GetInputString("type")
	var unit time.Duration
	s, ok := ctx.GetInputString("unit")
	if !ok {
		s = "s"
	}
	if u, err := getUnit(s); err != nil {
		return err
	} else {
		unit = u
	}

	switch strings.ToLower(t) {
	case "lognormal":
		mean, ok := ctx.GetInputFloat("mean")
		if !ok {
			return fmt.Errorf("missing required floating parameter 'mean'")
		}
		sigma, ok := ctx.GetInputFloat("mean")
		if !ok {
			return fmt.Errorf("missing required floating parameter 'mean'")
		}
		number := math.Round(random.NormFloat64()*sigma + mean)
		delay := time.Duration(number) * unit
		ctx.Log("Sleeping for %v", delay)
		time.Sleep(delay)
	case "uniform":
		lower, ok := ctx.GetInputInt("lower")
		if !ok {
			return fmt.Errorf("missing required floating parameter 'lower'")
		}
		upper, ok := ctx.GetInputInt("upper")
		if !ok {
			return fmt.Errorf("missing required parameter 'upper'")
		}
		number := rand.Intn(upper-lower) + lower
		delay := time.Duration(number) * unit
		ctx.Log("Sleeping for %v", delay)
		time.Sleep(delay)
	case "":
		t, _ := ctx.GetInput("time")
		var delay time.Duration
		switch t := t.(type) {
		case string:
			d, err := time.ParseDuration(t)
			if err != nil {
				return err
			}
			delay = d
		case float64:
			delay = time.Duration(t) * unit
		}
		ctx.Log("Sleeping for %v", delay)
		time.Sleep(delay)
	}

	return nil
}

func getUnit(unit string) (time.Duration, error) {
	switch strings.ToLower(unit) {
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
		return 0, errors.Errorf("unknown unit '%v'", unit)
	}
}
