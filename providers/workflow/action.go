package workflow

import (
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/workflow/actions"
	"mokapi/providers/workflow/event"
	"mokapi/providers/workflow/functions"
	"mokapi/providers/workflow/runtime"
)

var (
	actionCollection = map[string]runtime.Action{
		"xpath":     &actions.XPath{},
		"read-file": &actions.ReadFile{},
		"parse-yml": &actions.YmlParser{},
		"mustache":  &actions.Mustache{},
		"split":     &actions.Split{},
	}
	fCollection = map[string]functions.Function{
		"find": functions.Find,
	}
)

func RegisterAction(name string, action runtime.Action) {
	actionCollection[name] = action
}

func Run(workflows []mokapi.Workflow, event event.Handler, options ...WorkflowOptions) *runtime.Summary {
	summary := &runtime.Summary{}

	for _, w := range workflows {
		for _, trigger := range w.On {
			if event(trigger) {
				ctx := runtime.NewWorkflowContext(actionCollection, fCollection)
				for _, opt := range options {
					opt(ctx)
				}

				s, _ := runtime.Run(w, ctx)
				summary.Workflows = append(summary.Workflows, s)
			}
		}
	}

	return summary
}
