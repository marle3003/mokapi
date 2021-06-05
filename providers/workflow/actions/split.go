package actions

import (
	"fmt"
	"mokapi/providers/workflow/runtime"
	"strings"
)

type Split struct{}

func (m *Split) Run(ctx *runtime.ActionContext) error {
	s, ok := ctx.GetInputString("s")
	if !ok {
		return fmt.Errorf("missing required parameter 'template'")
	}
	separator, ok := ctx.GetInputString("separator")
	if !ok {
		return fmt.Errorf("missing required parameter 'data'")
	}
	n := -1
	count, ok := ctx.GetInput("n")
	if ok {
		if i, ok := count.(int); ok {
			n = i
		}
	}

	values := strings.SplitN(s, separator, n)

	for i, v := range values {
		ctx.SetOutput(fmt.Sprintf("_%v", i), v)
	}

	return nil
}
