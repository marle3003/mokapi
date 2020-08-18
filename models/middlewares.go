package models

import "time"

type ReplaceContent struct {
	Replacement Replacement
	Regex       string
}

type Replacement struct {
	From     string
	Selector string
}

type FilterContent struct {
	Filter *Filter
}

type Template struct {
	Filename string
}

type Selection struct {
	Slice *Slice
	First bool
}

type Slice struct {
	Low  int
	High int
}

type Delay struct {
	Distribution DelayDistribution
	Fixed        time.Duration
}

type DelayDistribution struct {
	Type   string
	Median float64
	Sigma  float64
	Lower  int
	Upper  int
}
