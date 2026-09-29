package generator

import (
	"fmt"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

type resolver struct {
	history []*schema.Schema
}

func resolve(req *Request, fallback bool) (*faker, error) {
	r := resolver{}
	return r.resolve(req, fallback)
}

func (r *resolver) resolve(req *Request, fallback bool) (*faker, error) {
	if len(req.Path) == 0 && req.Schema != nil && req.Schema.Id != "" {
		u, _ := url.Parse(req.Schema.Id)
		if u != nil {
			req.Path = append(req.Path, strings.Split(u.Path, "/")...)
		}
	}

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
		inferType := inferTypeFromKeywords(s)
		switch {
		case s.IsObject() && s.IsArray():
			n := gofakeit.Number(0, 1)
			if n == 0 {
				return r.resolveObject(req)
			}
			return r.resolveArray(req)
		case s.IsObject() || inferType == "object":
			return r.resolveObject(req)
		case s.IsArray() || inferType == "array":
			return r.resolveArray(req)
		default:
			switch {
			case len(s.AnyOf) > 0:
				i := gofakeit.Number(0, len(s.AnyOf)-1)
				return r.resolve(req.WithSchema(s.AnyOf[i]), fallback)
			case len(s.AllOf) > 0:
				allOf, err := intersectSchemas(s.AllOf...)
				if err != nil {
					return nil, fmt.Errorf("generate random data for schema failed: %w", err)
				}
				return r.resolve(req.WithSchema(allOf), fallback)
			case len(s.OneOf) > 0:
				return r.oneOf(req)
			}
		}

	}

	if s == nil && len(req.Path) > 0 {
		last := req.Path[len(req.Path)-1]
		if isPlural(last) {
			return r.resolve(req.With(req.Path, &schema.Schema{Type: schema.Types{"array"}}, req.examples), true)
		}
	}

	path := tokenize(req.Path)
	n := findBestMatch(g.root, req.WithPath(path))
	if n == g.root.defaultNode {
		n = findBestMatch(g.root, req)
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
			return root.defaultNode
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
		for _, attr := range child.Attributes {
			if attr == "*" || strings.EqualFold(attr, token) {
				match := child.findBestMatch(r.shift())
				if match != nil {
					return match
				}
			}
		}
	}

	if len(r.Path) > 1 {
		singular := g.inflector.Singular(token)
		if singular != token {
			r.Path[0] = singular
			match := n.findBestMatch(r)
			if match != nil {
				return match
			}
		}
	}

	// merge path
	if len(r.Path) > 1 {
		var sb strings.Builder
		sb.WriteString(r.Path[0])
		for _, p := range r.Path[1:] {
			sb.WriteString(p)
		}
		merged := r.WithPath([]string{sb.String()})
		match := n.findBestMatch(merged)
		if match != nil {
			return match
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
			if arr, ok := v.([]any); ok {
				if len(arr) > 0 {
					if _, err := p.Parse(arr[0]); err == nil {
						return newFaker(func() (any, error) {
							i := r.g.rand.Intn(len(arr))
							return arr[i], nil
						}), true
					}
				}
			}
		}
	}
	return nil, false
}
