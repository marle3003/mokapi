package mokapi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

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

func (triggers *Triggers) UnmarshalYAML(n *yaml.Node) error {
	m := make(map[string]interface{})
	err := n.Decode(m)
	if err != nil {
		return err
	}

	for k, v := range m {
		switch key := strings.ToLower(k); {
		case key == "service":
			*triggers = append(*triggers, parseTrigger(v)...)
		case key == "http":
			for _, h := range parseHttpTriggers(v) {
				*triggers = append(*triggers, Trigger{Http: h})
			}
		case key == "schedule":
			*triggers = append(*triggers, Trigger{Schedule: parseSchedule(v)})
		}
	}

	return nil
}

func parseHttpTriggers(i interface{}) (t []*HttpTrigger) {
	switch i := i.(type) {
	case []interface{}:
		for _, m := range i {
			t = append(t, &HttpTrigger{Method: fmt.Sprintf("%v", m)})
		}
	case map[string]interface{}:
		for k, v := range i {
			switch key := strings.ToLower(k); {
			case key == "path":
				switch v := v.(type) {
				case string:
					t = append(t, &HttpTrigger{Path: v})
				case []interface{}:
					for _, p := range v {
						t = append(t, &HttpTrigger{Path: fmt.Sprintf("%v", p)})
					}
				}

			case key == "get", key == "post":
				t = append(t, parseEndpoint(key, v)...)
			}
		}
	}

	return
}

func parseSchedule(i interface{}) (t *ScheduleTrigger) {
	t = &ScheduleTrigger{}
	switch i := i.(type) {
	case map[string]interface{}:
		for k, v := range i {
			s := fmt.Sprintf("%v", v)
			switch key := strings.ToLower(k); {
			case key == "every":
				t.Every = s
			case key == "iterations":
				if i, err := strconv.Atoi(s); err != nil {
					log.Errorf("error parsing int %q", v)
				} else {
					t.Iterations = i
				}
			}
		}
	}
	return
}

func parseTrigger(i interface{}) (t Triggers) {
	switch i := i.(type) {
	case map[string]interface{}:
		name := ""
		var http []*HttpTrigger
		for k, v := range i {
			switch k {
			case "name":
				name = fmt.Sprintf("%v", v)
			case "http":
				http = parseHttpTriggers(v)
			}
		}
		if len(http) == 0 {
			t = append(t, Trigger{Service: name})
		} else {
			for _, h := range http {
				t = append(t, Trigger{Service: name, Http: h})
			}
		}
	}
	return
}

func parseEndpoint(method string, i interface{}) (t []*HttpTrigger) {
	if i == nil {
		t = append(t, &HttpTrigger{Method: method})
	} else {
		switch i := i.(type) {
		case string:
			t = append(t, &HttpTrigger{Method: method, Path: i})
		case []interface{}:
			for _, p := range i {
				t = append(t, &HttpTrigger{Method: method, Path: fmt.Sprintf("%v", p)})
			}
		case map[string]interface{}:
			for k, v := range i {
				switch strings.ToLower(k) {
				case "path":
					if paths, ok := v.([]interface{}); ok {
						for _, path := range paths {
							t = append(t, &HttpTrigger{Method: method, Path: fmt.Sprintf("%s", path)})
						}
					} else {
						t = append(t, &HttpTrigger{Method: method, Path: fmt.Sprintf("%s", v)})
					}
				}
			}
		}
	}

	return
}
