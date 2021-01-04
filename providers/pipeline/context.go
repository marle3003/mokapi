package pipeline

import (
	"mokapi/providers/pipeline/types"
)

type contextModifier func(ctx *context) error

type context struct {
	outer *context
	steps map[string]Step
	vars  map[string]types.Object
}

func newContext(modifiers ...contextModifier) *context {
	c := &context{steps: map[string]Step{}, vars: map[string]types.Object{}}
	for _, m := range modifiers {
		err := m(c)
		if err != nil {
			panic(err)
		}
	}
	return c
}

func withVars(vars map[string]types.Object) contextModifier {
	return func(c *context) error {
		for k, v := range vars {
			c.vars[k] = v
		}
		return nil
	}
}

func withOuter(outer *context) contextModifier {
	return func(c *context) error {
		c.outer = outer
		return nil
	}
}

func (c *context) Get(t Type) interface{} {
	key := string(t)
	if v, ok := c.vars[key]; ok {
		return v.(*types.Reference).Value()
	}

	if c.outer != nil {
		return c.outer.Get(t)
	}

	return nil
}

func (c *context) Set(t Type, i interface{}) error {
	key := string(t)
	obj, err := types.Convert(i)
	if err != nil {
		return err
	}
	c.vars[key] = obj
	return nil
}

func (c *context) getVar(name string) (types.Object, bool) {
	if v, ok := c.vars[name]; ok {
		return v, true
	}
	if c.outer != nil {
		return c.outer.getVar(name)
	}
	return nil, false
}

func (c *context) setVar(name string, value types.Object) {
	c.vars[name] = value
}

func (c *context) getStep(name string) (Step, bool) {
	if c.steps != nil {
		if cmd, ok := c.steps[name]; ok {
			return cmd, true
		}
	}
	if c.outer != nil {
		return c.outer.getStep(name)
	}

	return nil, false
}
