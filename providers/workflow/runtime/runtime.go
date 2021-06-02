package runtime

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/utils"
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/parser"
	"strings"
	"time"
)

func Run(workflow mokapi.Workflow, ctx *WorkflowContext) (*WorkflowSummary, error) {
	summary := &WorkflowSummary{Name: workflow.Name}
	start := time.Now()
	defer func() {
		end := time.Now()
		summary.Duration = end.Sub(start)
	}()

	for k, v := range workflow.Env {
		v, err := sPrint(v, ctx)
		if err != nil {
			return summary, err
		}
		err = ctx.SetEnv(k, v)
		if err != nil {
			return summary, err
		}
	}

	for _, step := range workflow.Steps {
		if stepSum, err := runStep(step, ctx); err != nil {
			log.Error(err)
		} else {
			summary.Steps = append(summary.Steps, stepSum)
		}
	}

	return summary, nil
}

func runStep(step mokapi.Step, ctx *WorkflowContext) (*StepSummary, error) {
	summary := NewStepSummary(step)
	start := time.Now()

	ctx.OpenScope()

	defer func() {
		end := time.Now()
		summary.Duration = end.Sub(start)
		ctx.CloseScope()
	}()

	for k, v := range step.Env {
		if err := ctx.SetEnv(k, v); err != nil {
			return summary, err
		}
	}

	var err error
	if len(step.Run) > 0 {
		summary.Log, err = runScript(step, ctx)
	} else {
		summary.Log, err = runAction(step, ctx)
	}

	return summary, err
}

func RunExpression(s string, ctx *WorkflowContext) (interface{}, error) {
	exp := parser.Parse(s)

	v := newVisitor(ctx)
	ast.Walk(v, exp)

	return v.stack.Pop(), nil
}

func runScript(step mokapi.Step, ctx *WorkflowContext) (log string, err error) {
	script := step.Run
	script, err = sPrint(script, ctx)
	if err != nil {
		return
	}

	var output []byte
	switch shell := step.Shell; {
	case len(shell) == 0:
		if ctx.GOOS == "windows" {
			output, err = runCmd(script, ctx)
		}
		output, err = runBash(script, ctx)
	}
	if err != nil {
		return
	}
	log = fmt.Sprintf("%s", output)
	ctx.parseOutput(log, step.Id)

	return
}

func runAction(step mokapi.Step, ctx *WorkflowContext) (log string, err error) {
	stepId := getStepId(step)
	ctx.Context.NewStep(stepId)

	for k, v := range step.With {
		var val interface{}
		val, err = parse(v, ctx)
		if err != nil {
			return
		}
		ctx.Context.Steps[stepId].Inputs[k] = val
	}

	if a, ok := ctx.Actions[step.Uses]; !ok {
		return "", fmt.Errorf("unknown action %v", step.Uses)
	} else {
		actionCtx := newActionContext(stepId, ctx)
		err = a.Run(actionCtx)
		log = strings.Join(actionCtx.log, "\n")
	}

	return
}

func getStepId(step mokapi.Step) string {
	if len(step.Id) > 0 {
		return step.Id
	}
	return utils.NewGuid()
}
