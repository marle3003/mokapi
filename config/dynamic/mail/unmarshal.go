package mail

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"regexp"
)

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type alias Config
	tmp := alias(*c)
	tmp.AutoCreateMailbox = true
	err := value.Decode(&tmp)
	if err != nil {
		return err
	}
	*c = Config(tmp)
	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	type alias Config
	tmp := alias(*c)
	tmp.AutoCreateMailbox = true
	err := dec.Decode(&tmp)
	if err != nil {
		return err
	}
	*c = Config(tmp)
	return nil
}

func (r *RuleExpr) UnmarshalYAML(value *yaml.Node) error {
	var err error
	r.expr, err = regexp.Compile(value.Value)
	return err
}
