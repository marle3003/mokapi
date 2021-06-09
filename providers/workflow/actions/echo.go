package actions

import (
	"fmt"
	"mokapi/providers/workflow/runtime"
)

type Echo struct {
}

func (e *Echo) Run(ctx *runtime.ActionContext) error {
	msg, ok := ctx.GetInputString("msg")
	if !ok {
		return fmt.Errorf("missing required parameter 'msg'")
	}

	ctx.Log(msg)

	return nil
}
