package runtime

import (
	"fmt"
)

type Action interface {
	Run(ctx *ActionContext) error
}

type ActionContext struct {
	ctx    *WorkflowContext
	stepId string
	log    []string
}

func newActionContext(stepId string, ctx *WorkflowContext) *ActionContext {
	return &ActionContext{
		ctx:    ctx,
		stepId: stepId,
	}
}

func (c *ActionContext) GetInput(name string) (interface{}, bool) {
	val, ok := c.ctx.Context.Steps[c.stepId].Inputs[name]
	return val, ok
}

func (c *ActionContext) GetInputString(name string) (string, bool) {
	v, found := c.GetInput(name)
	return fmt.Sprintf("%s", v), found
}

func (c *ActionContext) SetOutput(name string, value interface{}) {
	c.ctx.Context.Steps[c.stepId].Outputs[name] = value
}

func (c *ActionContext) WorkflowContext() *WorkflowContext {
	return c.ctx
}

func (c *ActionContext) Log(format string, args ...interface{}) {
	c.log = append(c.log, fmt.Sprintf(format, args...))
}

func (c *ActionContext) GetInputInt(name string) (int, bool) {
	val, ok := c.ctx.Context.Steps[c.stepId].Inputs[name]
	if !ok {
		return 0, false
	}
	i, ok := val.(int)
	return i, ok
}

func (c *ActionContext) GetInputFloat(name string) (float64, bool) {
	val, ok := c.ctx.Context.Steps[c.stepId].Inputs[name]
	if !ok {
		return 0, false
	}
	f, ok := val.(float64)
	return f, ok
}
