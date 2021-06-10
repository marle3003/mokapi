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

	for k, v := range workflow.Vars {
		v, err := parse(v, ctx)
		if err != nil {
			return summary, err
		}
		ctx.Context.Set(k, v)
	}

	for _, step := range workflow.Steps {
		stepSum, err := runStep(step, ctx)
		summary.Steps = append(summary.Steps, stepSum)
		if err != nil {
			if stepSum != nil {
				stepSum.Status = Error
			}
			summary.Status = Error
			log.Error(err)
			return summary, err
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

	if len(step.If) > 0 {
		expr := step.If
		// ${{ }} is optionally
		if strings.HasPrefix(expr, "${{") && strings.HasSuffix(expr, "}}") {
			expr = expr[3 : len(expr)-2]
		}
		i, err := RunExpression(expr, ctx)
		if err != nil {
			return summary, err
		}
		if b, ok := i.(bool); !ok {
			return summary, fmt.Errorf("action id %q, if condition; expected bool value, got %t", summary.Id, i)
		} else if !b {
			return summary, nil
		}
	}

	for k, v := range step.Env {
		if err := ctx.SetEnv(k, v); err != nil {
			return summary, err
		}
	}

	ctx.Context.NewStep(summary.Id)

	var err error
	if len(step.Run) > 0 {
		summary.Log, err = runScript(step, summary.Id, ctx)
	} else {
		summary.Log, err = runAction(step, summary.Id, ctx)
	}

	if len(summary.Log) == 0 && err != nil {
		summary.Log = err.Error()
	}

	return summary, err
}

func RunExpression(s string, ctx *WorkflowContext) (interface{}, error) {
	exp := parser.Parse(s)

	v := newVisitor(ctx)
	ast.Walk(v, exp)

	return v.stack.Pop(), nil
}

func runScript(step mokapi.Step, stepId string, ctx *WorkflowContext) (s string, err error) {
	script := step.Run
	script, err = sPrint(script, ctx)
	if err != nil {
		return
	}

	var output []byte
	switch step.Shell {
	case "bash":
		output, err = runBash(script, ctx)
	case "ps":
		output, err = runPowershell(script, ctx)
	case "sh":
		output, err = runShell(script, ctx)
	case "cmd":
		output, err = runCmd(script, ctx)
	default:
		if strings.Contains(ctx.GOOS, "windows") {
			output, err = runCmd(script, ctx)
		} else {
			output, err = runBash(script, ctx)
		}
	}
	if err != nil {
		return
	}
	s = fmt.Sprintf("%s", output)
	ctx.parseOutput(s, stepId)

	return
}

func runAction(step mokapi.Step, stepId string, ctx *WorkflowContext) (s string, err error) {
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
		s = strings.Join(actionCtx.log, "\n")
	}

	return
}

func getStepId(step mokapi.Step) string {
	if len(step.Id) > 0 {
		return step.Id
	}
	return utils.NewGuid()
}
