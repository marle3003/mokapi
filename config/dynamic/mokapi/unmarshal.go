package mokapi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

func (triggers *Triggers) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		var s string
		err := n.Decode(&s)
		if err != nil {
			return err
		}
		switch s {
		case "http":
			*triggers = append(*triggers, Trigger{Http: &HttpTrigger{}})
		}
	} else if n.Kind == yaml.MappingNode {

		m := make(map[string]interface{})
		err := n.Decode(m)
		if err != nil {
			return err
		}

		for k, v := range m {
			switch key := strings.ToLower(k); {
			case key == "http":
				for _, h := range parseHttpTriggers(v) {
					*triggers = append(*triggers, Trigger{Http: h})
				}
			case key == "smtp":
				for _, s := range parseSmtp(v) {
					*triggers = append(*triggers, Trigger{Smtp: s})
				}
			case key == "schedule":
				*triggers = append(*triggers, Trigger{Schedule: parseSchedule(v)})
			}
		}
	}

	return nil
}

func parseHttpTriggers(i interface{}) (t []*HttpTrigger) {
	if i == nil {
		return []*HttpTrigger{{}}
	}
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

			case key == "get", key == "post", key == "put", key == "patch", key == "delete", key == "head", key == "options", key == "trace":
				t = append(t, parseEndpoint(key, v)...)
			}
		}
	case string:
		t = append(t, &HttpTrigger{Method: i})
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

func parseSmtp(i interface{}) (t []*SmtpTrigger) {
	if i == nil {
		return []*SmtpTrigger{{}}
	}
	switch i := i.(type) {
	case []interface{}:
		for _, m := range i {
			if s, ok := m.(string); ok {
				switch strings.ToLower(s) {
				case "login":
					t = append(t, &SmtpTrigger{Login: true})
				case "received":
					t = append(t, &SmtpTrigger{Received: true})
				case "logout":
					t = append(t, &SmtpTrigger{Logout: true})
				}
			}
		}
	case map[string]interface{}:
		var addr string
		for k, v := range i {
			switch key := strings.ToLower(k); {
			case key == "login":
				t = append(t, &SmtpTrigger{Login: v.(bool)})
			case key == "received":
				t = append(t, &SmtpTrigger{Received: v.(bool)})
			case key == "logout":
				t = append(t, &SmtpTrigger{Logout: v.(bool)})
			case key == "address":
				addr = v.(string)
			}
		}
		if len(addr) > 0 {
			for _, ti := range t {
				ti.Address = addr
			}
		}
	case string:
		switch strings.ToLower(i) {
		case "login":
			t = append(t, &SmtpTrigger{Login: true})
		case "received":
			t = append(t, &SmtpTrigger{Received: true})
		case "logout":
			t = append(t, &SmtpTrigger{Logout: true})
		}
	}

	return
}
