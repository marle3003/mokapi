package runtime

import (
	"fmt"
	"mokapi/providers/workflow/functions"
	"regexp"
	"runtime"
	"strings"
)

var (
	outputPattern = regexp.MustCompile(`::[^ ]+\s+name=(?P<name>\w+)::(?P<value>\w*)`)
)

type WorkflowContext struct {
	env              *envScope
	GOOS             string
	Output           strings.Builder
	Context          *Context
	Actions          map[string]Action
	Functions        map[string]functions.Function
	WorkingDirectory string
}

type envScope struct {
	parent *envScope
	env    map[string]interface{}
}

func NewWorkflowContext(actions map[string]Action, functions map[string]functions.Function) *WorkflowContext {
	return &WorkflowContext{
		env: &envScope{
			env: make(map[string]interface{}),
		},
		GOOS:      runtime.GOOS,
		Context:   newContext(),
		Actions:   actions,
		Functions: functions,
	}
}

func (ctx *WorkflowContext) ParseOutput(output []byte, stepId string) {
	s := fmt.Sprintf("%s", output)
	ctx.Output.WriteString(s)
	matches := outputPattern.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		name := m[1]
		value := m[2]
		ctx.Set(fmt.Sprintf("steps.%v.%v", stepId, name), value)
	}
}

func (ctx *WorkflowContext) OpenScope() {
	env := &envScope{
		parent: ctx.env,
		env:    make(map[string]interface{}),
	}
	ctx.env = env
}

func (ctx *WorkflowContext) CloseScope() {
	ctx.env = ctx.env.parent
}

func (ctx *WorkflowContext) EnvStrings() []string {
	return ctx.env.envStrings()
}

func (ctx *WorkflowContext) Get(name string) interface{} {
	return ctx.env.get(name)
}

func (ctx *WorkflowContext) GetString(name string) string {
	v := ctx.env.get(name)
	if v == nil {
		return ""
	}

	return fmt.Sprintf("%s", v)
}

func (ctx *WorkflowContext) Set(name string, value interface{}) {
	ctx.env.set(name, value)
}

func (ctx *WorkflowContext) ResolvePath(path string) string {
	if strings.HasPrefix(path, "./") {
		return strings.Replace(path, ".", ctx.WorkingDirectory, 1)
	}

	return path
}

func (ctx *envScope) get(name string) interface{} {
	if val, ok := ctx.env[name]; ok {
		return val
	}
	if ctx.parent != nil {
		return ctx.parent.get(name)
	}
	return nil
}

func (ctx *envScope) set(name string, value interface{}) {
	ctx.env[name] = value
}

func (ctx *envScope) envStrings() []string {
	r := make([]string, 0)
	if ctx.parent != nil {
		r = append(r, ctx.parent.envStrings()...)
	}

	for k, v := range ctx.env {
		r = append(r, fmt.Sprintf("%v=%v", k, v))
	}
	return r
}
