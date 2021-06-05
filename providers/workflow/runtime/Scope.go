package runtime

import (
	"mokapi/providers/workflow/functions"
	"os"
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
	ctx := &WorkflowContext{
		GOOS:      runtime.GOOS,
		Context:   newContext(),
		Actions:   actions,
		Functions: functions,
	}
	ctx.OpenScope()
	for _, v := range os.Environ() {
		kv := strings.SplitN(v, "=", 2)
		ctx.Env.set(kv[0], kv[1])
	}
	return ctx
}

func (ctx *WorkflowContext) parseOutput(s string, stepId string) {
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
		env:    make(map[string]string),
	}
	ctx.Env = env
	ctx.Context.data["env"] = env
}

func (ctx *WorkflowContext) CloseScope() {
	ctx.Env = ctx.Env.parent
	ctx.Context.data["env"] = ctx.Env
}

func (ctx *WorkflowContext) Environ() []string {
	return ctx.Env.environ()
}

func (ctx *WorkflowContext) ResolvePath(path string) string {
	if strings.HasPrefix(path, "./") {
		return strings.Replace(path, ".", ctx.WorkingDirectory, 1)
	}

	return path
}

func (ctx *WorkflowContext) GetEnv(name string) string {
	return ctx.Env.get(name)
}

func (ctx *WorkflowContext) SetEnv(name, value string) error {
	s, err := sPrint(value, ctx)
	if err != nil {
		return err
	}
	ctx.Env.set(name, s)

	return nil
}
