package models

import "mokapi/providers/parser"

type Middleware struct {
	ReplaceContent *ReplaceContent
	FilterContent  *FilterContent
}

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
