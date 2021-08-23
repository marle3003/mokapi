package mokapi

//func (w *Workflow) Run(ctx *runtime.WorkflowContext) (*runtime.WorkflowSummary, error) {
//	summary := &runtime.WorkflowSummary{Name: w.Name}
//	start := time.Now()
//	defer func() {
//		end := time.Now()
//		summary.Duration = end.Sub(start)
//	}()
//
//	for k, v := range w.Env {
//		err := ctx.SetEnv(k, v)
//		if err != nil {
//			return summary, err
//		}
//	}
//
//	for k, v := range w.Vars {
//		err := ctx.SetCtx(k, v)
//		if err != nil {
//			return summary, err
//		}
//	}
//
//	for _, step := range w.Steps {
//		stepSum, err := runStep(step, ctx)
//		summary.Steps = append(summary.Steps, stepSum)
//		if err != nil {
//			if stepSum != nil {
//				stepSum.Status = runtime.Error
//			}
//			summary.Status = runtime.Error
//			log.Error(err)
//			return summary, err
//		}
//	}
//
//	return summary, nil
//}
//
//func runStep(step Step, ctx *runtime.WorkflowContext) (*runtime.StepSummary, error) {
//	summary := newStepSummary(step)
//	start := time.Now()
//
//	ctx.OpenScope()
//
//	defer func() {
//		end := time.Now()
//		summary.Duration = end.Sub(start)
//		ctx.CloseScope()
//	}()
//
//	if len(step.If) > 0 {
//		expr := step.If
//		// ${{ }} is optionally
//		if strings.HasPrefix(expr, "${{") && strings.HasSuffix(expr, "}}") {
//			expr = expr[3 : len(expr)-2]
//		}
//		i, err := runtime.RunExpression(expr, ctx)
//		if err != nil {
//			return summary, err
//		}
//		if b, ok := i.(bool); !ok {
//			return summary, fmt.Errorf("action id %q, if condition; expected bool value, got %v", summary.Id, i)
//		} else if !b {
//			summary.Status = runtime.Skip
//			return summary, nil
//		}
//	}
//
//	for k, v := range step.Env {
//		if err := ctx.SetEnv(k, v); err != nil {
//			return summary, err
//		}
//	}
//
//	ctx.Context.NewStep(summary.Id)
//
//	var err error
//	if len(step.Run) > 0 {
//		summary.Log, err = runtime.RunScript(step.Run, step.Shell, summary.Id, ctx)
//	} else {
//		summary.Log, err = runtime.RunAction(step.Uses, step.With, summary.Id, ctx)
//	}
//
//	return summary, err
//}
//
//func getStepId(step Step) string {
//	if len(step.Id) > 0 {
//		return step.Id
//	}
//	return utils.NewGuid()
//}
//
//func newStepSummary(step Step) *runtime.StepSummary {
//	name := step.Name
//	if len(name) == 0 {
//		if len(step.Run) > 0 {
//			name = step.Run
//		} else {
//			name = step.Uses
//		}
//	}
//
//	return &runtime.StepSummary{Name: name, Id: getStepId(step), Log: make([]string, 0)}
//}
