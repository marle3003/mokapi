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
}

func NewActionContext(stepId string, ctx *WorkflowContext) *ActionContext {
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
