package events

import (
	"fmt"
	"strings"
)

const (
	namespace string = "namespace"
	name      string = "name"
)

type Traits map[string]string

func NewTraits() Traits {
	return make(Traits)
}

func (t Traits) WithNamespace(ns string) Traits {
	return t.With(namespace, ns)
}

func (t Traits) WithName(v string) Traits {
	return t.With(name, v)
}

func (t Traits) GetName() string {
	return t.Get(name)
}

func (t Traits) With(name, value string) Traits {
	t[name] = value
	return t
}

func (t Traits) Get(name string) string {
	return t[name]
}

func (t Traits) String() string {
	var sb strings.Builder
	if ns, ok := t[namespace]; ok {
		sb.WriteString("namespace=" + ns)
	}
	if n, ok := t[name]; ok {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("name=" + n)
	}
	for k, v := range t {
		if k == namespace || k == name {
			continue
		}
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
