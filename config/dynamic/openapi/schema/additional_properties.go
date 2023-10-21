package schema

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type AdditionalProperties struct {
	*Ref
	Forbidden bool
}

func (ap *AdditionalProperties) IsFreeForm() bool {
	if ap == nil {
		return true
	}
	if ap.Ref == nil || ap.Value == nil {
		return !ap.Forbidden
	}
	if ap.Value != nil && ap.Value.Type == "" {
		return true
	}
	return false
}

func (ap *AdditionalProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		var b bool
		err := node.Decode(&b)
		if err != nil {
			return err
		}
		ap.Forbidden = !b
		return err
	} else {
		return node.Decode(&ap.Ref)
	}
}

func (ap *AdditionalProperties) UnmarshalJSON(b []byte) error {
	var allowed bool
	err := json.Unmarshal(b, &allowed)
	if err == nil {
		ap.Forbidden = !allowed
		return nil
	}
	return json.Unmarshal(b, &ap.Ref)
}
