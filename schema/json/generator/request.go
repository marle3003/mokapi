package generator

import "mokapi/schema/json/schema"

type Request struct {
	Path   []string       `json:"path"`
	Schema *schema.Schema `json:"schema"`

	g        *generator
	ctx      *context
	examples []any
}

func NewRequest(path []string, s *schema.Schema, ctx map[string]any) *Request {
	return &Request{
		Path:   path,
		Schema: s,
		ctx:    &context{store: ctx},
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
		ctx:      r.ctx,
		examples: example,
	}
}

type context struct {
	store
	snapshots []store
}

type store map[string]any

func newContext() *context {
	return &context{store: make(store)}
}

func (c *context) Snapshot() {
	c.snapshots = append(c.snapshots, c.store.Snapshot())
}

func (c *context) Restore() {
	snapshot := c.snapshots[len(c.snapshots)-1]
	c.snapshots = c.snapshots[:len(c.snapshots)-1]
	c.store = snapshot
}

func (s *store) Snapshot() store {
	snapshot := map[string]any{}
	for k, v := range *s {
		snapshot[k] = v
	}
	return snapshot
}
