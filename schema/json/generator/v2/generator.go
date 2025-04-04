package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"mokapi/schema/json/schema"
	"regexp"
	"strings"
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

	g *generator
}

var g = &generator{
	rand: rand.New(rand.NewSource(time.Now().Unix())),
	root: buildTree(),
}

func Seed(seed int64) {
	g.rand.Seed(seed)
}

func New(r *Request) (interface{}, error) {
	r.g = g
	r.tokenizePath()
	v, err := fakeWalk(g.root, r)
	if err == nil {
		return v, nil
	}
	return fakeBySchema(r)
}

func fakeBySchema(r *Request) (interface{}, error) {
	if fake, ok := applyConstraints(r); ok {
		return fake()
	}

	s := r.Schema
	if s.IsString() {
		return fakeString(r)
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

func (r *Request) tokenizePath() {
	var result []string
	for _, path := range r.Path {
		result = append(result, splitWords(path)...)
	}
	r.Path = result
}

func (r *Request) NextToken() string {
	if len(r.Path) == 0 {
		return ""
	}
	return r.Path[0]
}

// splitWords splits camelCase and dot notation into words
func splitWords(s string) []string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, "${1} ${2}")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ToLower(s)
	return strings.Fields(s)
}
