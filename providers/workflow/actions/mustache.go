package actions

import (
	"fmt"
	"mokapi/providers/workflow/path/objectpath"
	"mokapi/providers/workflow/runtime"
	"regexp"
	"strings"
)

var pattern = regexp.MustCompile(`{{\s*([\w\.]+)\s*}}`)

type Mustache struct{}

func (m *Mustache) Run(ctx *runtime.ActionContext) error {
	template, ok := ctx.GetInputString("template")
	if !ok {
		return fmt.Errorf("missing required parameter 'template'")
	}
	data, ok := ctx.GetInput("data")
	if !ok {
		return fmt.Errorf("missing required parameter 'data'")
	}

	matches := pattern.FindAllStringSubmatch(template, -1) // matches is [][]string

	for _, match := range matches {
		path := strings.TrimSpace(match[1])
		var value interface{}
		var err error
		if path == "." {
			value = data
		} else {
			value, err = objectpath.Resolve(path, data)
			if err != nil {
				return err
			}
		}

		template = strings.ReplaceAll(template, match[0], fmt.Sprintf("%v", value))
	}

	ctx.SetOutput("result", template)

	return nil
}
