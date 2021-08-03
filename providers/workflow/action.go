package workflow

import (
	"fmt"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/workflow/actions"
	"mokapi/providers/workflow/event"
	"mokapi/providers/workflow/functions"
	"mokapi/providers/workflow/runtime"
)

var (
	actionCollection = map[string]runtime.Action{
		"xpath":      &actions.XPath{},
		"read-file":  &actions.ReadFile{},
		"parse-yml":  &actions.YmlParser{},
		"parse-yaml": &actions.YmlParser{},
		"mustache":   &actions.Mustache{},
		"split":      &actions.Split{},
		"echo":       &actions.Echo{},
		"delay":      &actions.Delay{},
		"send-mail":  &actions.SendMail{},
	}
	fCollection = map[string]functions.Function{
		"find":      functions.Find,
		"findAll":   functions.FindAll,
		"any":       functions.Any,
		"format":    functions.Format,
		"now":       functions.Now,
		"randInt":   functions.RandInt,
		"randFloat": functions.RandFloat,
		"hasPrefix": functions.HasPrefix,
		"hasSuffix": functions.HasSuffix,
		"contains":  functions.Contains,
		"toLower":   functions.ToLower,
		"toUpper":   functions.ToUpper,
	}
)

func RegisterAction(name string, action runtime.Action) {
	actionCollection[name] = action
}

func Run(workflows []mokapi.Workflow, event event.Handler, options ...Options) (*runtime.Summary, error) {
	summary := &runtime.Summary{}

	for _, w := range workflows {
		for _, trigger := range w.On {
			if event(trigger) {
				ctx := runtime.NewWorkflowContext(actionCollection, fCollection)
				for _, opt := range options {
					opt(ctx)
				}

				s, err := runtime.Run(w, ctx)
				if err != nil {
					return nil, fmt.Errorf("workflow %v: %v", w.Name, err.Error())
				}
				summary.Workflows = append(summary.Workflows, s)
			}
		}
	}

	return summary, nil
}
