package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type resolver struct {
	history []*schema.Schema
}

func resolve(req *Request, fallback bool) (*faker, error) {
	r := resolver{}
	return r.resolve(req, fallback)
}

func (r *resolver) resolve(req *Request, fallback bool) (*faker, error) {
	if f, ok := useFromContext(req); ok {
		return f, nil
	}

	if fake, ok := applyConstraints(req); ok {
		return newFaker(fake), nil
	}

	s := req.Schema
	if f, ok := nullable(s); ok {
		return f, nil
	}

	if err := r.guardLoopLimit(s); err != nil {
		if s.IsNullable() {
			return nullFaker, nil
		}
		return nil, err
	}
	if s != nil {
		r.history = append(r.history, s)
		defer func() {
			if len(r.history) == 0 || s == nil {
				return
			}
			r.history = r.history[:len(r.history)-1]
		}()
	}

	if s != nil {
		switch {
		case len(s.AnyOf) > 0:
			i := gofakeit.Number(0, len(s.AnyOf)-1)
			return r.resolve(req.WithSchema(s.AnyOf[i]), fallback)
		case len(s.AllOf) > 0:
			return r.allOf(req)
		case len(s.OneOf) > 0:
			return r.oneOf(req)
		}

		if s.IsObject() || s.HasProperties() {
			return r.resolveObject(req)
		} else if s.IsArray() {
			return r.resolveArray(req)
		}
	}

	path := tokenize(req.Path)
	if s == nil {
		if len(path) > 0 {
			last := path[len(path)-1]
			if isPlural(last) {
				return r.resolve(req.With(path, &schema.Schema{Type: schema.Types{"array"}}, req.examples), true)
			}
		}
	}
	n := findBestMatch(g.root, req.With(path, req.Schema, req.examples))
	if n == nil {
		if !fallback {
			return nil, NoMatchFound
		}
		return newFaker(func() (any, error) {
			return fakeBySchema(req)
		}), nil
	}
	return newFakerWithFallback(n, req), nil
}

func (r *resolver) guardLoopLimit(s *schema.Schema) error {
	if s == nil {
		return nil
	}
	// recursion guard. Currently, we use a fixed depth: 1
	numRequestsSameAsThisOne := 0
	for _, h := range r.history {
		if s == h {
			numRequestsSameAsThisOne++
		} else if s.Ref != "" && s.Ref == h.Ref {
			numRequestsSameAsThisOne++
		}
	}
	if numRequestsSameAsThisOne >= 1 {
		return &RecursionGuard{s: s}
	}
	return nil
}

func findBestMatch(root *Node, r *Request) *Node {
	for {
		if match := root.findBestMatch(r); match != nil {
			return match
		}
		if len(r.Path) == 0 {
			return nil
		}
		r = r.shift()
	}
}

func (n *Node) findBestMatch(r *Request) *Node {
	token := r.NextToken()
	if token == "" && n.Name != "root" && n.Fake != nil {
		return n
	}

	for _, child := range n.Children {
		var attributes = child.Attributes
		if len(attributes) == 0 {
			attributes = []string{child.Name}
		}
		for _, attr := range attributes {
			if attr == token || attr == "*" {
				match := child.findBestMatch(r.shift())
				if match != nil {
					return match
				}
			}
		}
	}

	// Check if the current token exists in the root
	for _, child := range g.root.Children {
		if child.Name == token {
			return nil
		}
		for _, attr := range child.Attributes {
			if attr == token {
				return nil
			}
		}
	}

	// Skip current token
	skip := r.shift()
	if len(skip.Path) > 0 {
		return n.findBestMatch(skip)
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

func getPathFromRef(ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	return strings.ToLower(filepath.Base(u.Fragment))
}

// splitWords splits camelCase and dot notation into words
func splitWords(s string) []string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, "${1} ${2}")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ToLower(s)
	return strings.Fields(s)
}

type RecursionGuard struct {
	s *schema.Schema
}

func (e *RecursionGuard) Error() string {
	return fmt.Sprintf("recursion in object path found but schema does not allow null: %v", e.s)
}

func nullable(s *schema.Schema) (*faker, bool) {
	if s != nil && s.IsNullable() {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return newFaker(func() (any, error) {
				return nil, nil
			}), true
		}
	}
	return nil, false
}

func useFromContext(r *Request) (*faker, bool) {
	if len(r.Path) > 0 {
		last := r.Path[len(r.Path)-1]
		if v, ok := r.Context.Values[last]; ok {
			p := parser.Parser{Schema: r.Schema}
			if v, err := p.Parse(v); err == nil {
				return newFaker(func() (any, error) {
					return v, nil
				}), true
			}
		}
	}
	return nil, false
}
