package mokapi

import "gopkg.in/yaml.v3"

func (v *Variables) UnmarshalYAML(n *yaml.Node) error {

	if n.Kind == yaml.SequenceNode {
		var s []Variable
		err := n.Decode(&s)
		if err != nil {
			return err
		}
		for _, i := range s {
			*v = append(*v, i)
		}
	} else if n.Kind == yaml.MappingNode {
		m := make(map[string]string)
		err := n.Decode(m)
		if err != nil {
			return err
		}
		for k, i := range m {
			*v = append(*v, Variable{Name: k, Value: i})
		}
	}

	return nil
}

func (v *Variable) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.MappingNode {
		m := make(map[string]string)
		err := n.Decode(m)
		if err != nil {
			return err
		}
		if len(m) == 2 {
			v.Name = m["name"]
			v.Value = m["value"]
		} else {
			// should only have one entry
			for name, value := range m {
				v.Name = name
				v.Value = value
			}
		}
	} else if n.Kind == yaml.ScalarNode {
		var s string
		n.Decode(&s)
		v.Expr = s
	}

	return nil
}
