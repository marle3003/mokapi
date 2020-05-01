package dynamic

import "strings"

type Ldap struct {
	Server  LdapServer
	Entries []*Entry
}

type LdapServer struct {
	Listen string
	Config map[string][]string
}

type Entry struct {
	Dn         string
	Attributes map[string][]string
}

type AttributeValueList struct {
	Values []string
}

func (s *LdapServer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s.Config = make(map[string][]string)

	data := make(map[string]string)
	unmarshal(data)

	for k, v := range data {
		if strings.ToLower(k) == "listen" {
			s.Listen = v
		} else {
			list := make([]string, 1)
			list[0] = v
			s.Config[k] = list
		}
	}

	listData := make(map[string][]string)
	unmarshal(listData)

	for k, v := range listData {
		s.Config[k] = v
	}

	return nil
}

func (e *Entry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	e.Attributes = make(map[string][]string)

	data := make(map[string]string)
	unmarshal(data)

	for k, v := range data {
		if strings.ToLower(k) == "dn" {
			e.Dn = v
			parts := strings.Split(v, ",")
			if len(parts) > 0 {
				kv := strings.Split(parts[0], "=")
				if len(kv) == 2 && strings.ToLower(kv[0]) == "cn" {
					e.Attributes["cn"] = []string{strings.TrimSpace(kv[1])}
				}
			}
		} else {
			list := make([]string, 1)
			list[0] = v
			e.Attributes[k] = list
		}
	}

	listData := make(map[string][]string)
	unmarshal(listData)

	for k, v := range listData {
		e.Attributes[k] = v
	}

	return nil
}
