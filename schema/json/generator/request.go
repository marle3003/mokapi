package generator

import (
	"maps"
	"mokapi/schema/json/schema"
	"slices"
)

type Request struct {
	Path    []string       `json:"path"`
	Schema  *schema.Schema `json:"schema"`
	Context *Context       `json:"context"`

	g        *generator
	examples []any
	// trace is only set when the request is explained, see Explain
	trace *trace
	// at is the path of the value in the result. In contrast to Path it is not rewritten
	// while matching nodes.
	at []string
}

func NewRequest(path []string, s *schema.Schema, ctx map[string]any) *Request {
	if ctx == nil {
		ctx = make(map[string]any)
	}

	return &Request{
		Path:    nonNilPath(path),
		Schema:  s,
		Context: &Context{Values: ctx},
	}
}

// nonNilPath keeps Request.Path an empty array instead of null in JSON and JavaScript.
// goja also converts a nil slice inside a map to null.
func nonNilPath(path []string) []string {
	if path == nil {
		return []string{}
	}
	return path
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
		Path:     nonNilPath(path),
		Schema:   s,
		g:        r.g,
		Context:  r.Context,
		examples: example,
		trace:    r.trace,
		at:       r.at,
	}
}

// property returns the request for a property or the items of an array below this request.
// It is only needed to explain a request, otherwise r is returned.
func (r *Request) property(name string) *Request {
	if r.trace == nil {
		return r
	}
	r2 := *r
	r2.at = append(slices.Clone(r.at), name)
	return &r2
}

type Context struct {
	Values map[string]any `json:"values"`
	// snapshots holds the values of the enclosing objects, the innermost last
	snapshots []map[string]any
}

type Values map[string]any

func newContext() *Context {
	return &Context{Values: make(Values)}
}

// Snapshot starts a nested level that sees a copy of the current values.
// Restore discards everything the nested level has written.
func (c *Context) Snapshot() {
	c.snapshots = append(c.snapshots, c.Values)
	c.Values = maps.Clone(c.Values)
}

func (c *Context) Restore() {
	if len(c.snapshots) == 0 {
		return
	}

	last := len(c.snapshots) - 1
	c.Values = c.snapshots[last]
	c.snapshots = c.snapshots[:last]
}

// forgetNode drops the value a node has stored under its name, so a nested level generates it
// again instead of reusing the one of its parent.
func (c *Context) forgetNode(n *Node) {
	if n != nil {
		delete(c.Values, n.Name)
	}
}

func (c *Context) Has(key string) bool {
	_, ok := c.Values[key]
	return ok
}

func (c *Context) Value(key string) any {
	return c.Values[key]
}
