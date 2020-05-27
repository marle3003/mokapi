package models

import "mokapi/providers/parser"

type ReplaceContent struct {
	Replacement Replacement
	Regex       string
}

type Replacement struct {
	From     string
	Selector string
}

type FilterContent struct {
	Filter *parser.FilterExp
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
