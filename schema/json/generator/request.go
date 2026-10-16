package generator

import (
	"maps"
	"mokapi/schema/json/schema"
)

type Request struct {
	Path    []string       `json:"path"`
	Schema  *schema.Schema `json:"schema"`
	Context *Context       `json:"context"`

	g        *generator
	examples []any
}

func NewRequest(path []string, s *schema.Schema, ctx map[string]any) *Request {
	if ctx == nil {
		ctx = make(map[string]any)
	}

	return &Request{
		Path:    path,
		Schema:  s,
		Context: &Context{Values: ctx},
	}
}

func (r *Request) shift() *Request {
	r2 := *r
	if len(r2.Path) > 0 {
		r2.Path = r2.Path[1:]
	}
	return &r2
}

func (r *Request) NextToken() string {
	if len(r.Path) == 0 {
		return ""
	}
	return r.Path[0]
}

func (r *Request) WithSchema(s *schema.Schema) *Request {
	return r.With(r.Path, s, r.examples)
}

func (r *Request) WithPath(path []string) *Request {
	return r.With(path, r.Schema, r.examples)
}

func (r *Request) With(path []string, s *schema.Schema, example []any) *Request {
	return &Request{
		Path:     path,
		Schema:   s,
		g:        r.g,
		Context:  r.Context,
		examples: example,
	}
}

type Context struct {
	Values    map[string]any `json:"values"`
	snapshots []map[string]any
	// represents how many steps current context is away from the root
	level int
}

type Values map[string]any

func newContext() *Context {
	return &Context{Values: make(Values)}
}

func (c *Context) Snapshot() {
	if c.level != 0 {
		m := map[string]any{}
		maps.Copy(m, c.Values)
		c.Values = map[string]any{}
		c.snapshots = append(c.snapshots, m)
	}
	c.level++
}

func (c *Context) Restore() {
	if len(c.snapshots) <= 1 {
		return
	}

	snapshot := c.snapshots[len(c.snapshots)-1]
	c.snapshots = c.snapshots[:len(c.snapshots)-1]
	c.Values = snapshot
}

func (c *Context) Has(key string) bool {
	_, ok := c.Values[key]
	return ok
}

func (c *Context) Value(key string) any {
	return c.Values[key]
}
