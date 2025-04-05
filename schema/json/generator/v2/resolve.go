package v2

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/schema"
	"regexp"
	"strings"
)

type resolver struct {
	history []*schema.Schema
}

func resolve(path []string, s *schema.Schema, fallback bool) (*faker, error) {
	r := resolver{}
	req := &Request{Path: path, Schema: s, ctx: map[string]interface{}{}, g: g}
	return r.resolve(req, fallback)
}

func (r *resolver) resolve(req *Request, fallback bool) (*faker, error) {
	s := req.Schema

	if err := r.guardLoopLimit(s); err != nil {
		if s.IsNullable() {
			return nullFaker, nil
		}
		return nil, err
	}
	r.history = append(r.history, s)
	defer func() {
		r.history = r.history[:len(r.history)-1]
	}()

	if s.IsObject() || s.HasProperties() {
		return r.resolveObject(req)
	} else if s.IsArray() {
		last := req.Path[len(req.Path)-1]
		item, err := r.resolve(req.With(append(req.Path, inflection.Singular(last)), s.Items), true)
		if err != nil {
			return nil, err
		}
		return newFaker(func() (interface{}, error) {
			list := &Request{
				Path:   req.Path,
				Schema: s,
			}
			return fakeArray(list, item)
		}), nil
	} else {
		path := tokenize(req.Path)
		if s == nil {
			last := path[len(path)-1]
			if isPlural(last) {
				return r.resolve(req.With(path, &schema.Schema{Type: schema.Types{"array"}}), true)
			}
		}
		n := findBestMatch(g.root, req.With(path, req.Schema))
		if n == nil && !fallback {
			return nil, NoMatchFound
		}
		return newFakerWithFallback(n, req), nil
	}
}

func (r *resolver) guardLoopLimit(s *schema.Schema) error {
	// recursion guard. Currently, we use a fixed depth: 1
	numRequestsSameAsThisOne := 0
	for _, h := range r.history {
		if s == h {
			numRequestsSameAsThisOne++
		}
	}
	if numRequestsSameAsThisOne > 1 {
		return &RecursionGuard{s: s}
	}
	return nil
}

func findBestMatch(root *Node, r *Request) *Node {
	for {
		if len(r.Path) == 0 {
			return nil
		}
		if match := root.findBestMatch(r); match != nil {
			return match
		}
		r = r.shift()
	}
}

func (n *Node) findBestMatch(r *Request) *Node {
	token := r.NextToken()
	if token == "" {
		return n
	}

	for _, child := range n.Children {
		if child.Name == token {
			match := child.findBestMatch(r.shift())
			if match != nil {
				return match
			}
		}
	}

	return nil
}

func tokenize(path []string) []string {
	var result []string
	for _, p := range path {
		result = append(result, splitWords(p)...)
	}
	return result
}

// splitWords splits camelCase and dot notation into words
func splitWords(s string) []string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, "${1} ${2}")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ToLower(s)
	return strings.Fields(s)
}

type RecursionGuard struct {
	s *schema.Schema
}

func (e *RecursionGuard) Error() string {
	return fmt.Sprintf("recursion in object path found but schema does not allow null: %v", e.s)
}
