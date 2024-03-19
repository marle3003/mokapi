package generator

import (
	"mokapi/json/schema"
	"net/url"
	"path/filepath"
)

type Path []*PathElement

type PathElement struct {
	Name   string      `json:"name"`
	Schema *schema.Ref `json:"schema"`
}

type Request struct {
	Path Path `json:"path"`

	g       *generator
	history []*schema.Ref
	context map[string]interface{}
}

func NewRequest(opts ...RequestOption) *Request {
	r := &Request{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Request) With(opts ...RequestOption) *Request {
	r1 := &Request{
		Path:    r.Path,
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
	last := r.Last()
	if last == nil {
		return ""
	}
	return last.Name
}

func (r *Request) LastSchema() *schema.Schema {
	last := r.Last()
	if last == nil || last.Schema == nil {
		return nil
	}
	return last.Schema.Value
}

func (r *Request) Last() *PathElement {
	if len(r.Path) == 0 {
		return nil
	}
	return r.Path[len(r.Path)-1]
}

//func (p Path) Last(c Comparer) bool {
//	if len(p) == 0 {
//		return false
//	}
//	return c.Compare(p[len(p)-1])
//}

//func (r *Request) GetName(index int) string {
//	if len(r.Names) == 0 {
//		return ""
//	}
//	if index < 0 {
//		index = len(r.Names) + index
//	}
//	if index < 0 || index >= len(r.Names) {
//		return ""
//	}
//	return r.Names[index]
//}

//func (r *Request) GetNames(index int) []string {
//	if len(r.Names) == 0 {
//		return nil
//	}
//	if index < 0 {
//		index = len(r.Names) + index
//	}
//	if index < 0 || index >= len(r.Names) {
//		return nil
//	}
//	return r.Names[index:]
//}

//func (p Path) Has(c Comparer) bool {
//	for _, n := range p {
//		if c.Compare(n) {
//			return true
//		}
//	}
//	return false
//}

type Comparer interface {
	Compare(p *PathElement) bool
}

type ComparerList []Comparer

func (p Path) MatchLast(list ...Comparer) bool {
	n1 := len(list)
	n2 := len(p)
	if n2 == 0 || n1 > n2 {
		return false
	}
	for i, c := range list {
		index := n2 - n1 + i
		if index >= n2 {
			return false
		}
		if !c.Compare(p[index]) {
			return false
		}
	}
	return true
}

//
//func (r *Request) matchLast(names []string, ignoreCase bool) bool {
//	n1 := len(names)
//	n2 := len(r.Names)
//	if n2 == 0 || n1 > n2 {
//		return false
//	}
//	for i, name := range names {
//		index := n2 - n1 + i
//		if index >= n2 {
//			return false
//		}
//		if ignoreCase && strings.ToLower(name) != strings.ToLower(r.Names[index]) {
//			return false
//		} else if !ignoreCase && name != r.Names[index] {
//			return false
//		}
//	}
//	return true
//}

func (p *PathElement) RefName() string {
	return RefName(p.Schema)
}

func RefName(r *schema.Ref) string {
	if r == nil || r.Ref == "" {
		return ""
	}
	u, err := url.Parse(r.Ref)
	if err != nil {
		return ""
	}
	return filepath.Base(u.Fragment)
}

type RequestOption func(r *Request)

func UsePathElement(name string, schema *schema.Ref) RequestOption {
	return func(r *Request) {
		r.Path = append(r.Path, &PathElement{Name: name, Schema: schema})
		r.history = append(r.history, schema)
	}
}

func PathElements(elements ...*PathElement) RequestOption {
	return func(r *Request) {
		r.Path = append(r.Path, elements...)
	}
}

//func Name(name ...string) RequestOption {
//	return func(r *Request) {
//		r.Names = append(r.Names, name...)
//	}
//}
//

//func Schema(r *schema.Ref) RequestOption {
//	return func(r *Request) {
//		if r == nil {
//			return
//		}
//		r.Schema = r
//		r.history = append(r.history, s.Value)
//	}
//}
