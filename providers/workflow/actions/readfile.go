package actions

import (
	"fmt"
	"io/ioutil"
	"mokapi/providers/workflow/runtime"
)

type ReadFile struct {
}

func (r *ReadFile) Run(ctx *runtime.ActionContext) error {
	path, ok := ctx.GetInputString("path")
	if !ok {
		return fmt.Errorf("missing required parameter 'path'")
	}

	data, err := ioutil.ReadFile(ctx.WorkflowContext().ResolvePath(path))
	if err != nil {
		return err
	}

	ctx.SetOutput("content", string(data))

	return nil
}
