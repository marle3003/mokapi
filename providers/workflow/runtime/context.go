package runtime

import "fmt"

type Context struct {
	Steps map[string]*StepContext
	data  map[string]interface{}
}

func (c *Context) Set(name string, value interface{}) {
	c.data[name] = value
}

func (c *Context) Get(name string) (interface{}, error) {
	switch name {
	case "steps":
		return c.Steps, nil
	default:
		i, ok := c.data[name]
		if !ok {
			return nil, fmt.Errorf("identifier %q not exists in current context", name)
		}
		return i, nil
	}
}

func (c *Context) NewStep(id string) {
	c.Steps[id] = newStepContext()
}

func newContext() *Context {
	return &Context{
		Steps: make(map[string]*StepContext),
		data:  make(map[string]interface{}),
	}
}

type StepContext struct {
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}

func (s *StepContext) Resolve(name string) (interface{}, error) {
	switch name {
	case "inputs":
		return s.Inputs, nil
	case "outputs":
		return s.Outputs, nil
	default:
		return nil, fmt.Errorf("unknown field '%q'", name)
	}
}

func newStepContext() *StepContext {
	return &StepContext{
		Inputs:  make(map[string]interface{}),
		Outputs: make(map[string]interface{}),
	}
}
