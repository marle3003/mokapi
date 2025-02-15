package directory

import (
	"encoding/base64"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/ldap"
	"mokapi/sortedmap"
	"net/url"
	"slices"
	"strings"
)

type Ldif struct {
	Records []LdifRecord
}

type LdifRecord interface {
	Apply(entries *sortedmap.LinkedHashMap[string, Entry], s *Schema) error
	append(name string, value string)
}

func (l *Ldif) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	l.Records = nil
	s := strings.ReplaceAll(string(config.Raw), "\r\n", "\n")
	lines := strings.Split(s, "\n")

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
			case "modrdn":
				rec = &ModifyDnRecord{Dn: dn, DeleteOldDn: true}
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

func (add *AddRecord) Apply(entries *sortedmap.LinkedHashMap[string, Entry], s *Schema) error {
	if _, ok := entries.Get(add.Dn); ok {
		return NewEntryError(ldap.EntryAlreadyExists, "record dn='%s' already exists", add.Dn)
	}
	e := Entry{
		Dn:         add.Dn,
		Attributes: add.Attributes,
	}
	err := e.validate(s)
	if err != nil {
		return err
	}
	entries.Set(e.Dn, e)
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

func (del *DeleteRecord) Apply(entries *sortedmap.LinkedHashMap[string, Entry], _ *Schema) error {
	_, ok := entries.Get(del.Dn)
	if !ok {
		return NewEntryError(ldap.NoSuchObject, "apply delete record failed: entry '%v' not found", del.Dn)
	}
	entries.Del(del.Dn)
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

func (mod *ModifyRecord) Apply(entries *sortedmap.LinkedHashMap[string, Entry], s *Schema) error {
	e, ok := entries.Get(mod.Dn)
	if !ok {
		return NewEntryError(ldap.NoSuchObject, "apply change record failed: entry '%v' not found", mod.Dn)
	}
	e = e.copy()

	if e.Attributes == nil {
		e.Attributes = make(map[string][]string)
	}

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
				if len(action.Attributes[action.Name]) == 0 {
					delete(e.Attributes, action.Name)
				} else {
					values := e.Attributes[action.Name]
					for _, val := range action.Attributes[action.Name] {
						values = slices.DeleteFunc(values, func(s string) bool {
							return s == val
						})
					}
					if len(values) == 0 {
						delete(e.Attributes, action.Name)
					} else {
						e.Attributes[action.Name] = values
					}
				}
			}
		case "replace":
			e.Attributes[action.Name] = action.Attributes[action.Name]
		}
	}

	if err := e.validate(s); err != nil {
		return err
	}
	entries.Set(e.Dn, e)

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

type ModifyDnRecord struct {
	Dn            string
	NewRdn        string
	DeleteOldDn   bool
	NewSuperiorDn string
}

func (m *ModifyDnRecord) Apply(entries *sortedmap.LinkedHashMap[string, Entry], _ *Schema) error {
	e, ok := entries.Get(m.Dn)
	if !ok {
		return NewEntryError(ldap.NoSuchObject, "apply modify DN failed: entry '%v' not found", m.Dn)
	}

	parts := strings.Split(e.Dn, ",")
	if m.NewRdn != "" {
		parts[0] = m.NewRdn
	}

	if m.NewSuperiorDn != "" {
		parts = append(parts[0:1], m.NewSuperiorDn)
	}

	e.Dn = strings.Join(parts, ",")

	if m.DeleteOldDn {
		entries.Del(m.Dn)
	}

	entries.Set(e.Dn, e)

	return nil
}

func (m *ModifyDnRecord) Validate(s *Schema) error {
	return nil
}

func (m *ModifyDnRecord) append(name string, value string) {
	switch name {
	case "newrdn":
		m.NewRdn = value
	case "deleteoldrdn":
		m.DeleteOldDn = value == "1"
	case "newsuperior":
		m.NewSuperiorDn = value
	}
}
