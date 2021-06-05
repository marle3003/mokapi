package workflow

import (
	"mokapi/providers/workflow/runtime"
)

type WorkflowOptions func(ctx *runtime.WorkflowContext)

func WithContext(name string, value interface{}) WorkflowOptions {
	return func(ctx *runtime.WorkflowContext) {
		ctx.Context.Set(name, value)
	}
}

func WithAction(name string, action runtime.Action) WorkflowOptions {
	return func(ctx *runtime.WorkflowContext) {
		ctx.Actions[name] = action
	}
}

func WithWorkingDirectory(path string) WorkflowOptions {
	return func(ctx *runtime.WorkflowContext) {
		ctx.WorkingDirectory = path
	}
}
