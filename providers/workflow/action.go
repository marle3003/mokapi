package workflow

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/utils"
	"mokapi/providers/workflow/actions"
	"mokapi/providers/workflow/event"
	"mokapi/providers/workflow/functions"
	"mokapi/providers/workflow/runtime"
	"regexp"
	"strings"
	"time"
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

type Handler func(events event.Handler, options ...runtime.WorkflowOptions)

func Run(action mokapi.Workflow, options ...runtime.WorkflowOptions) runtime.Summary {
	ctx := runtime.NewWorkflowContext(actionCollection, fCollection)
	summary := runtime.Summary{}
	start := time.Now()
	for _, o := range options {
		o(ctx)
	}

	for k, v := range action.Env {
		// TODO error
		v, _ := format(v, ctx)
		ctx.Env.Set(k, v)
	}

	for _, step := range action.Steps {
		if stepSum, err := runStep(step, ctx); err != nil {
			log.Error(err)
		} else {
			summary.Steps = append(summary.Steps, stepSum)
		}
	}
	end := time.Now()
	summary.Duration = end.Sub(start)
	log.WithField("log", summary).Debugf("Action %v", action.Name)

	return summary
}

func runStep(step mokapi.Step, ctx *runtime.WorkflowContext) (runtime.StepSummary, error) {
	summary := runtime.StepSummary{Name: step.Name}
	start := time.Now()
	defer func() {
		end := time.Now()
		summary.Duration = end.Sub(start)
	}()
	stepId := step.Id
	if len(stepId) == 0 {
		stepId = utils.NewGuid()
	}

	ctx.OpenScope()
	defer ctx.CloseScope()

	for k, v := range step.Env {
		// TODO error
		v, _ := format(v, ctx)
		ctx.Env.Set(k, v)
	}

	if len(step.Run) > 0 {
		if len(summary.Name) == 0 {
			summary.Name = step.Run
		}
		var output []byte
		var err error
		parsed, err := format(step.Run, ctx)
		if err != nil {
			return summary, err
		}
		s := fmt.Sprintf("%v", parsed)
		switch shell := step.Shell; {
		case len(shell) == 0:
			if ctx.GOOS == "windows" {
				output, err = runCmd(s, ctx)
			}
			output, err = runBash(s, ctx)
		}
		if err != nil {
			return summary, err
		}
		summary.Log = fmt.Sprintf("%s", output)
		ctx.ParseOutput(summary.Log, step.Id)
	} else if len(step.Uses) > 0 {
		if len(summary.Name) == 0 {
			summary.Name = step.Uses
		}
		ctx.Context.NewStep(stepId)
		for k, v := range step.With {
			val, err := format(v, ctx)
			if err != nil {
				return summary, err
			}
			ctx.Context.Steps[stepId].Inputs[k] = val
		}
		if a, ok := ctx.Actions[step.Uses]; ok {
			if err := a.Run(runtime.NewActionContext(stepId, ctx)); err != nil {
				return summary, err
			}
		} else {
			return summary, fmt.Errorf("unknown action %v", step.Uses)
		}
	} else {
		return summary, fmt.Errorf("unable to run step %q", summary.Name)
	}

	return summary, nil
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
