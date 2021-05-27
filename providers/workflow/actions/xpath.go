package actions

import (
	"gopkg.in/xmlpath.v2"
	"mokapi/providers/workflow/runtime"
	"strings"
)

type XPath struct {
}

func (x *XPath) Run(ctx *runtime.ActionContext) error {
	exp, _ := ctx.GetInputString("expression")
	s, _ := ctx.GetInputString("content")

	p := xmlpath.MustCompile(exp)
	root, err := xmlpath.Parse(strings.NewReader(s))
	if err != nil {
		return err
	}
	if val, ok := p.String(root); ok {
		ctx.SetOutput("result", val)
	} else {
		ctx.SetOutput("result", "")
	}

	return nil
}
