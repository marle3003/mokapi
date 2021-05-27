package runtime

type WorkflowOptions func(ctx *WorkflowContext)

func WithContext(name string, value interface{}) WorkflowOptions {
	return func(ctx *WorkflowContext) {
		ctx.Context.Set(name, value)
	}
}

func WithAction(name string, action Action) WorkflowOptions {
	return func(ctx *WorkflowContext) {
		ctx.Actions[name] = action
	}
}

func WithWorkingDirectory(path string) WorkflowOptions {
	return func(ctx *WorkflowContext) {
		ctx.WorkingDirectory = path
	}
}
