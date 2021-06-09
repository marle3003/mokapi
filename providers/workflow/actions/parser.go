package actions

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/providers/workflow/runtime"
)

type YmlParser struct{}

func (p *YmlParser) Run(ctx *runtime.ActionContext) error {
	content, ok := ctx.GetInputString("content")
	if !ok {
		return fmt.Errorf("missing required parameter 'content'")
	}

	result := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(content), result); err != nil {
		return err
	}

	ctx.SetOutput("result", result)

	return nil
}
