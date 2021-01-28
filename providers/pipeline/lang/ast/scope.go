package ast

import (
	"mokapi/providers/pipeline/lang/types"
)

type Scope struct {
	Outer   *Scope
	symbols map[string]types.Object
}

func NewScope(symbols map[string]types.Object) *Scope {
	base := &Scope{symbols: map[string]types.Object{"true": types.NewBool(true), "false": types.NewBool(false)}}
	outer := &Scope{symbols: symbols, Outer: base}
	return OpenScope(outer)
}

func OpenScope(outer *Scope) *Scope {
	return &Scope{Outer: outer, symbols: make(map[string]types.Object)}
}

func NewScopeWithOuter(symbols map[string]types.Object, outer *Scope) *Scope {
	return &Scope{Outer: outer, symbols: symbols}
}

func (c *Scope) Symbol(name string) (types.Object, bool) {
	if v, ok := c.symbols[name]; ok {
		return v, true
	}
	if c.Outer != nil {
		return c.Outer.Symbol(name)
	}
	return nil, false
}

func (c *Scope) SetSymbol(name string, val types.Object) {
	c.symbols[name] = val
}

func (c *Scope) Get(t types.Type) interface{} {
	key := string(t)
	if v, ok := c.symbols[key]; ok {
		return v.(*types.Reference).Elem()
	}

	if c.Outer != nil {
		return c.Outer.Get(t)
	}

	return nil
}
