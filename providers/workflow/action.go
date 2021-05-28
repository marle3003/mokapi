package workflow

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/utils"
	"mokapi/providers/workflow/actions"
	"mokapi/providers/workflow/functions"
	"mokapi/providers/workflow/runtime"
	"os/exec"
	"regexp"
	"strings"
)

var (
	actionCollection = map[string]runtime.Action{
		"xpath":     &actions.XPath{},
		"read-file": &actions.ReadFile{},
		"parse-yml": &actions.YmlParser{},
	}
	fCollection = map[string]functions.Function{
		"find": functions.Find,
	}
)

func Run(action mokapi.Workflow, options ...runtime.WorkflowOptions) {
	ctx := runtime.NewWorkflowContext(actionCollection, fCollection)
	for _, o := range options {
		o(ctx)
	}

	for k, v := range action.Env {
		// TODO error
		v, _ := format(v, ctx)
		ctx.Set(k, v)
	}

	for _, step := range action.Steps {
		if err := runStep(step, ctx); err != nil {
			log.Error(err)
		}
	}
	log.Debugf("Action %v: %v", action.Name, ctx.Output.String())
}

func runStep(step mokapi.Step, ctx *runtime.WorkflowContext) error {
	stepId := step.Id
	if len(stepId) == 0 {
		stepId = utils.NewGuid()
	}

	ctx.OpenScope()
	defer ctx.CloseScope()

	for k, v := range step.Env {
		// TODO error
		v, _ := format(v, ctx)
		ctx.Set(k, v)
	}

	if len(step.Run) > 0 {
		switch shell := step.Shell; {
		case len(shell) == 0:
			if ctx.GOOS == "windows" {
				return runCmd(step, ctx)
			}
			return runBash(step, ctx)
		}
	} else if len(step.Uses) > 0 {
		ctx.Context.NewStep(stepId)
		for k, v := range step.With {
			val, err := format(v, ctx)
			if err != nil {
				return err
			}
			ctx.Context.Steps[stepId].Inputs[k] = val
		}
		if a, ok := ctx.Actions[step.Uses]; ok {
			return a.Run(runtime.NewActionContext(stepId, ctx))
		} else {
			return fmt.Errorf("unknown action %v", step.Uses)
		}
	}

	name := step.Name
	if len(name) == 0 {
		name = step.Run
	}
	return fmt.Errorf("unable to run step %q", name)
}

func runBash(step mokapi.Step, ctx *runtime.WorkflowContext) error {
	path, err := exec.LookPath("bash")
	if err != nil {
		return runShell(step, ctx)
	}

	cmd := exec.Command(path, "-c", step.Run)
	cmd.Env = ctx.EnvStrings()

	output, err := cmd.Output()
	if err != nil {
		return err
	}
	ctx.ParseOutput(output, step.Id)

	return nil
}

func runShell(step mokapi.Step, ctx *runtime.WorkflowContext) error {
	path, err := exec.LookPath("sh")
	if err != nil {
		return errors.Wrap(err, "unable to run step")
	}

	cmd := exec.Command(path, "-c", step.Run)
	cmd.Env = ctx.EnvStrings()

	output, err := cmd.Output()
	if err != nil {
		return err
	}
	ctx.ParseOutput(output, step.Id)

	return nil
}

func runCmd(step mokapi.Step, ctx *runtime.WorkflowContext) error {
	path, err := exec.LookPath("cmd")
	if err != nil {
		return fmt.Errorf("cmd not found")
	}
	cmd := &exec.Cmd{
		Path: path,
		Args: []string{"/C", step.Run},
		Env:  ctx.EnvStrings(),
	}

	output, err := cmd.Output()
	if err != nil {
		return err
	}
	ctx.ParseOutput(output, step.Id)

	return nil
}

func format(s string, ctx *runtime.WorkflowContext) (interface{}, error) {
	if strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}") {
		return runtime.RunExpression(s[3:len(s)-2], ctx)
	}

	p := regexp.MustCompile(`\${{(?P<exp>[^}]*)}}`)
	matches := p.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		i, err := runtime.RunExpression(m[1], ctx)
		if err != nil {
			return s, err
		}
		s = strings.Replace(s, m[0], fmt.Sprintf("%v", i), 1)
	}

	return s, nil
}
