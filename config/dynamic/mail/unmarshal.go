package mail

import (
	"gopkg.in/yaml.v3"
	"regexp"
)

func (r *RuleExpr) UnmarshalYAML(value *yaml.Node) error {
	var err error
	r.expr, err = regexp.Compile(value.Value)
	return err
}
