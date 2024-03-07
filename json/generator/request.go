package generator

import (
	"mokapi/json/schema"
	"strings"
)

type Request struct {
	Names  []string
	Schema *schema.Schema

	g       *generator
	history []*schema.Schema
	context map[string]interface{}
}

func (r *Request) With(opts ...RequestOption) *Request {
	r1 := &Request{
		Names:   r.Names,
		Schema:  r.Schema,
		g:       r.g,
		history: r.history,
		context: r.context,
	}
	for _, opt := range opts {
		opt(r1)
	}
	return r1
}

func (r *Request) LastName() string {
	if len(r.Names) == 0 {
		return ""
	}
	return r.Names[len(r.Names)-1]
}

func (r *Request) GetName(index int) string {
	if len(r.Names) == 0 {
		return ""
	}
	if index < 0 {
		index = len(r.Names) + index
	}
	if index < 0 || index >= len(r.Names) {
		return ""
	}
	return r.Names[index]
}

func (r *Request) GetNames(index int) []string {
	if len(r.Names) == 0 {
		return nil
	}
	if index < 0 {
		index = len(r.Names) + index
	}
	if index < 0 || index >= len(r.Names) {
		return nil
	}
	return r.Names[index:]
}

func (r *Request) matchLast(names []string, ignoreCase bool) bool {
	n1 := len(names)
	n2 := len(r.Names)
	if n2 == 0 || n1 > n2 {
		return false
	}
	for i, name := range names {
		index := n2 - n1 + i
		if index >= n2 {
			return false
		}
		if ignoreCase && strings.ToLower(name) != strings.ToLower(r.Names[index]) {
			return false
		} else if !ignoreCase && name != r.Names[index] {
			return false
		}
	}
	return true
}

type RequestOption func(r *Request)

func Name(name ...string) RequestOption {
	return func(r *Request) {
		r.Names = append(r.Names, name...)
	}
}

func Schema(s *schema.Schema) RequestOption {
	return func(r *Request) {
		r.Schema = s
		r.history = append(r.history, s)
	}
}

func Ref(s *schema.Ref) RequestOption {
	return func(r *Request) {
		if s == nil {
			return
		}
		r.Schema = s.Value
		r.history = append(r.history, s.Value)
	}
}
