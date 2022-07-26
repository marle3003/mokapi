package events

import (
	"fmt"
	"strings"
)

type Traits map[string]string

func NewTraits() Traits {
	return make(Traits)
}

func (t Traits) WithNamespace(ns string) Traits {
	return t.With("namespace", ns)
}

func (t Traits) WithName(name string) Traits {
	return t.With("name", name)
}

func (t Traits) With(name, value string) Traits {
	t[name] = value
	return t
}

func (t Traits) String() string {
	var sb strings.Builder
	for k, v := range t {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v=%v", k, v))
	}
	return sb.String()
}

func (t Traits) Match(traits Traits) bool {
	for k, v := range t {
		if s, ok := traits[k]; !ok || s != v {
			return false
		}
	}
	return true
}

func (t Traits) Contains(traits Traits) bool {
	for k, v := range traits {
		if s, ok := t[k]; !ok || s != v {
			return false
		}
	}
	return true
}
