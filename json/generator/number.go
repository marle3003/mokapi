package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
)

const (
	Int32   NumberFormat = iota
	Int64   NumberFormat = 1
	Float32 NumberFormat = 2
	Float64 NumberFormat = 3
)

type NumberFormat int

type NumberOptions struct {
	Format   NumberFormat
	Minimum  *float64
	Maximum  *float64
	Nullable bool
}

func NewNumber(opt NumberOptions) (interface{}, error) {
	if opt.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil, nil
		}
	}

	if opt.Minimum != nil && opt.Maximum != nil &&
		(*opt.Minimum) > (*opt.Maximum) {
		return nil, fmt.Errorf("invalid minimum '%v' and maximum '%v'", *opt.Minimum, *opt.Maximum)
	}

	hasRange := opt.Minimum != nil || opt.Maximum != nil

	switch opt.Format {
	case Int32:
		if !hasRange {
			return gofakeit.Int32(), nil
		}
		min := math.MinInt32
		max := math.MaxInt32
		if opt.Minimum != nil {
			min = int(math.Ceil(*opt.Minimum))
		}
		if opt.Maximum != nil {
			max = int(math.Floor(*opt.Maximum))
		}
		if err := validateRange(min, max); err != nil {
			return 0, err
		}

		// gofakeit uses Intn function which panics if number is <= 0
		return int32(math.Round(float64(gofakeit.Float32Range(float32(min), float32(max))))), nil
	case Int64:
		if !hasRange {
			return gofakeit.Int64(), nil
		}
		max := math.MaxInt64
		min := math.MinInt64
		if opt.Minimum != nil {
			min = int(math.Ceil(*opt.Minimum))
		}
		if opt.Maximum != nil {
			max = int(math.Floor(*opt.Maximum))
		}
		if err := validateRange(min, max); err != nil {
			return 0, err
		}

		return int64(math.Round(gofakeit.Float64Range(float64(min), float64(max)))), nil
	case Float32:
		if !hasRange {
			return gofakeit.Float32(), nil
		}
		max := float32(math.MaxFloat32)
		min := max * -1
		if opt.Minimum != nil {
			min = float32(*opt.Minimum)
		}
		if opt.Maximum != nil {
			max = float32(*opt.Maximum)
		}
		return gofakeit.Float32Range(min, max), nil
	case Float64:
		if !hasRange {
			return gofakeit.Float64(), nil
		}
		max := math.MaxFloat64
		min := max * -1
		if opt.Minimum != nil {
			min = *opt.Minimum
		}
		if opt.Maximum != nil {
			max = *opt.Maximum
		}
		return gofakeit.Float64Range(min, max), nil
	default:
		return 0, fmt.Errorf("format %v not supported", opt.Format)
	}
}

func validateRange(min, max int) error {
	if min >= max {
		return fmt.Errorf("invalid minimum '%v' and maximum '%v'", min, max)
	}
	return nil
}

func hasNumberRange(opt NumberOptions) bool {
	return opt.Minimum != nil || opt.Maximum != nil
}
