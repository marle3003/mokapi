package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"mokapi/schema/json/schema"
	"time"
)

var types = []string{"string", "number", "integer", "boolean", "array", "object"}

type generator struct {
	rand *rand.Rand

	root *Node
}

type Request struct {
	Path   []string `json:"path"`
	Schema *schema.Schema

	g   *generator
	ctx *context
}

var g = &generator{
	rand: rand.New(rand.NewSource(time.Now().Unix())),
	root: buildTree(),
}

func Seed(seed int64) {
	g.rand.Seed(seed)
}

func New(r *Request) (interface{}, error) {
	f, err := resolve(r.Path, r.Schema, true)
	if err != nil {
		return nil, err
	}
	return f.fake()
}

func fakeBySchema(r *Request) (interface{}, error) {
	if fake, ok := applyConstraints(r); ok {
		return fake()
	}

	s := r.Schema
	switch {
	case s.IsString():
		return fakeString(r)
	case s.IsObject():
		return fakeObject(r.Schema)
	case s.IsArray():
		items := func() (interface{}, error) {
			return fakeBySchema(r.WithSchema(s.Items))
		}
		return fakeArray(r, newFaker(items))
	case s.Is("boolean"):
		return gofakeit.Bool(), nil
	case s.IsNumber():
		return fakeNumber(r)
	case s.IsInteger():
		return fakeInteger(r.Schema)
	case s.IsNullable():
		return nil, nil
	}

	i := gofakeit.Number(0, len(types)-1)
	r.Schema = &schema.Schema{Type: schema.Types{types[i]}}
	return fakeBySchema(r)
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
	return r.With(r.Path, s)
}

func (r *Request) WithPath(path []string) *Request {
	return r.With(path, r.Schema)
}

func (r *Request) With(path []string, s *schema.Schema) *Request {
	return &Request{
		Path:   path,
		Schema: s,
		g:      r.g,
		ctx:    r.ctx,
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
