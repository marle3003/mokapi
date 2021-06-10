package workflow

import (
	"mokapi/providers/workflow/runtime"
)

type Options func(ctx *runtime.WorkflowContext)

func WithContext(name string, value interface{}) Options {
	return func(ctx *runtime.WorkflowContext) {
		ctx.Context.Set(name, value)
	}
}

func WithAction(name string, action runtime.Action) Options {
	return func(ctx *runtime.WorkflowContext) {
		ctx.Actions[name] = action
	}
}

func WithWorkingDirectory(path string) Options {
	return func(ctx *runtime.WorkflowContext) {
		ctx.WorkingDirectory = path
	}
}
