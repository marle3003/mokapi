package runtime

import (
	"mokapi/providers/pipeline/lang/types"
)

type Scope struct {
	outer   *Scope
	symbols map[string]types.Object
}

func NewScope(symbols map[string]types.Object) *Scope {
	outer := &Scope{symbols: symbols}
	return &Scope{outer: outer, symbols: make(map[string]types.Object)}
}

func NewScopeWithOuter(symbols map[string]types.Object, outer *Scope) *Scope {
	return &Scope{outer: outer, symbols: symbols}
}

func (c *Scope) Symbol(name string) (types.Object, bool) {
	if v, ok := c.symbols[name]; ok {
		return v, true
	}
	if c.outer != nil {
		return c.outer.Symbol(name)
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

	if c.outer != nil {
		return c.outer.Get(t)
	}

	return nil
}
