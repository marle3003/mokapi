package directory

import (
	"encoding/base64"
	"fmt"
	"mokapi/config/dynamic"
	"net/url"
	"strings"
)

type Ldif struct {
	Records []LdifRecord
}

type LdifRecord interface {
	Apply(entries map[string]Entry) error
	append(name string, value string)
}

func (l *Ldif) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	l.Records = nil
	lines := strings.Split(string(config.Raw), "\n")

	if len(lines) > 0 {
		kv := strings.SplitN(lines[0], ":", 2)
		if kv[0] == "version" {
			lines = lines[1:]
		}
	}

	trim := func(s string) string {
		if s == "" {
			return ""
		}
		if s[0] == ' ' {
			return s[1:]
		}
		return s
	}
	isMultiLine := func(next int) bool {
		if next < len(lines) {
			line := lines[next]
			if len(line) > 0 {
				return lines[next][0] == ' '
			}
		}
		return false
	}

	var rec LdifRecord
	var dn string
	parsedLine := ""
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			if rec != nil {
				l.Records = append(l.Records, rec)
				rec = nil
			}
			continue
		}
		if line[0] == '#' || line[0] == '-' {
			continue
		}

		if isMultiLine(i + 1) {
			parsedLine += trim(line)
			continue
		} else if parsedLine != "" {
			line = parsedLine + trim(line)
			parsedLine = ""
		}

		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid line %v: %s", i, line)
		}
		val := trim(kv[1])

		switch kv[0] {
		case "dn":
			dn = formatDN(val)
			rec = &AddRecord{Dn: dn}
		case "changetype":
			switch val {
			case "add":
				rec = &AddRecord{Dn: dn}
			case "delete":
				rec = &DeleteRecord{Dn: dn}
			case "modify":
				rec = &ModifyRecord{Dn: dn}
			default:
				return fmt.Errorf("changetype %s not supported", val)
			}
		default:
			if rec == nil {
				return fmt.Errorf("no DN set at line %v: %v", i, line)
			}

			if len(val) > 0 {
				switch val[0] {
				case '<':
					u, err := url.Parse(strings.TrimSpace(val[1:]))
					if err != nil {
						return fmt.Errorf("parse attribute %s failed: %w", kv[0], err)
					}
					var c *dynamic.Config
					c, err = reader.Read(u, nil)
					if err != nil {
						return fmt.Errorf("read file for attribute %s failed: %w", kv[0], err)
					}
					dynamic.AddRef(config, c)
					val = string(c.Raw)
				case ':':
					b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(val[1:]))
					if err != nil {
						return fmt.Errorf("base64 decode attribute %s failed: %w", kv[0], err)
					}
					val = string(b)
				}
			}

			rec.append(kv[0], val)
		}
	}
	if rec != nil {
		l.Records = append(l.Records, rec)
	}

	return nil
}

type AddRecord struct {
	Dn         string
	Attributes map[string][]string
}

func (add *AddRecord) Apply(entries map[string]Entry) error {
	if _, ok := entries[add.Dn]; ok {
		return fmt.Errorf("record dn='%s' already exists", add.Dn)
	}
	entries[add.Dn] = Entry{
		Dn:         add.Dn,
		Attributes: add.Attributes,
	}
	return nil
}

func (add *AddRecord) append(name string, value string) {
	if add.Attributes == nil {
		add.Attributes = make(map[string][]string)
	}
	add.Attributes[name] = append(add.Attributes[name], value)
}

type DeleteRecord struct {
	Dn string
}

func (del *DeleteRecord) Apply(entries map[string]Entry) error {
	return nil
}

func (del *DeleteRecord) append(_ string, _ string) {}

type ModifyRecord struct {
	Dn      string
	Actions []*ModifyAction
}

type ModifyAction struct {
	Type       string
	Name       string
	Attributes map[string][]string
}

func (mod *ModifyRecord) Apply(entries map[string]Entry) error {
	if _, ok := entries[mod.Dn]; !ok {
		return fmt.Errorf("apply change record failed: entry '%v' not found", mod.Dn)
	}
	e := entries[mod.Dn]

	for _, action := range mod.Actions {
		switch action.Type {
		case "add":
			if _, ok := e.Attributes[action.Name]; !ok {
				e.Attributes[action.Name] = action.Attributes[action.Name]
			} else {
				e.Attributes[action.Name] = append(e.Attributes[action.Name], action.Attributes[action.Name]...)
			}
		case "delete":
			if _, ok := e.Attributes[action.Name]; ok {
				delete(e.Attributes, action.Name)
			}
		case "replace":
			e.Attributes[action.Name] = action.Attributes[action.Name]
		}
	}

	return nil
}

func (mod *ModifyRecord) append(name string, value string) {
	switch name {
	case "add":
		mod.Actions = append(mod.Actions, &ModifyAction{
			Type: "add",
			Name: value,
		})
	case "replace":
		mod.Actions = append(mod.Actions, &ModifyAction{
			Type: "replace",
			Name: value,
		})
	case "delete":
		mod.Actions = append(mod.Actions, &ModifyAction{
			Type: "delete",
			Name: value,
		})
	default:
		if len(mod.Actions) == 0 {
			return
		}
		action := mod.Actions[len(mod.Actions)-1]
		if action.Attributes == nil {
			action.Attributes = make(map[string][]string)
		}
		action.Attributes[name] = append(action.Attributes[name], value)
	}
}
