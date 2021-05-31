package runtime

import (
	"mokapi/providers/workflow/functions"
	"regexp"
	"runtime"
	"strings"
)

var (
	outputPattern = regexp.MustCompile(`::[^ ]+\s+name=(?P<name>\w+)::(?P<value>\w*)`)
)

type WorkflowContext struct {
	Env              *Env
	GOOS             string
	Context          *Context
	Actions          map[string]Action
	Functions        map[string]functions.Function
	WorkingDirectory string
	Summary          Summary
}

func NewWorkflowContext(actions map[string]Action, functions map[string]functions.Function) *WorkflowContext {
	env := &Env{
		env: make(map[string]interface{}),
	}
	ctx := newContext()
	ctx.data["env"] = env
	return &WorkflowContext{
		Env:       env,
		GOOS:      runtime.GOOS,
		Context:   ctx,
		Actions:   actions,
		Functions: functions,
	}
}

func (ctx *WorkflowContext) ParseOutput(s string, stepId string) {
	matches := outputPattern.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		name := m[1]
		value := m[2]
		ctx.Context.Steps[stepId].Outputs[name] = value
	}
}

func (ctx *WorkflowContext) OpenScope() {
	env := &Env{
		parent: ctx.Env,
		env:    make(map[string]interface{}),
	}
	ctx.Env = env
	ctx.Context.data["env"] = env
}

func (ctx *WorkflowContext) CloseScope() {
	ctx.Env = ctx.Env.parent
	ctx.Context.data["env"] = ctx.Env
}

func (ctx *WorkflowContext) EnvStrings() []string {
	return ctx.Env.envStrings()
}

func (ctx *WorkflowContext) ResolvePath(path string) string {
	if strings.HasPrefix(path, "./") {
		return strings.Replace(path, ".", ctx.WorkingDirectory, 1)
	}

	return path
}
