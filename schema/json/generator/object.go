package generator

import (
	"errors"
	"fmt"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"regexp/syntax"
	"slices"
	"strings"
	"unicode"

	"github.com/brianvoe/gofakeit/v6"
)

var (
	numPatternProperties     = []interface{}{0, 1, 2, 3, 4, 5}
	weightsPatternProperties = []float32{0.1, 3, 2, 1, 0.5, 0.5}
)

func (r *resolver) resolveObject(req *Request) (*faker, error) {
	s := req.Schema

	if len(req.Path) == 0 {
		domain := detectDomain(s, g.root)
		if domain != "" {
			req.Path = append(req.Path, domain)
		}
	}

	var fakes *sortedmap.LinkedHashMap[string, *faker]
	var err error
	resetStore := false

	switch {
	case !s.HasProperties() && s.AdditionalProperties != nil:
		fakes, err = r.fakeDictionary(req)
		resetStore = true
	case !s.HasProperties() && s.PatternProperties == nil:
		if len(s.Examples) > 0 {
			return fakeByExample(req)
		}
		match := findBestMatch(g.root, req)
		return newFakerWithFallback(match, req), nil
	default:
		fakes, err = r.fakeObject(req)
	}

	if err != nil {
		return nil, err
	}

	fake := func() (interface{}, error) {
		m := map[string]interface{}{}
		var sorted []string
		sorted, err = topologicalSort(fakes)
		if err != nil {
			return nil, err
		}

		for _, key := range sorted {
			if resetStore {
				req.Context.Snapshot()
			}
			f := fakes.Lookup(key)
			m[key], err = f.fake()
			if err != nil {
				return nil, err
			}
			if resetStore {
				req.Context.Restore()
			}
		}

		if s.If != nil {
			p := parser.Parser{}
			_, err := p.ParseWith(m, s.If)
			var cond *schema.Schema
			if err == nil && s.Then != nil {
				cond = s.Then
			} else if err != nil && s.Else != nil {
				cond = s.Else
			}
			if cond != nil {
				f, err := r.resolve(req.WithSchema(cond), true)
				if err != nil {
					return nil, err
				}
				v, err := f.fake()
				if err != nil {
					return nil, err
				}
				if m2, ok := v.(map[string]interface{}); ok {
					for key, val := range m2 {
						m[key] = val
					}
				}
			}
		}

		return m, nil
	}

	return newFaker(fake), nil
}

func (r *resolver) fakeObject(req *Request) (*sortedmap.LinkedHashMap[string, *faker], error) {
	s := req.Schema
	fakes := &sortedmap.LinkedHashMap[string, *faker]{}
	domain := detectDomain(s, g.root)
	fallback := domain == ""
	req.examples = examplesFromRequest(req)
	if !isKnownDomain(req) {
		req.Path = append(req.Path, domain)
	}

	if s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			prop := append(req.Path, it.Key())
			ex := propertyFromExample(it.Key(), req)
			f, err := r.resolve(req.With(prop, it.Value(), ex), fallback)
			if err != nil {
				var guard *RecursionGuard
				if errors.As(err, &guard) {
					if !slices.Contains(req.Schema.Required, it.Key()) {
						continue
					}
				}
				if errors.Is(err, NoMatchFound) {
					if domain != "" {
						f, err = r.resolve(req.With([]string{domain, it.Key()}, it.Value(), ex), true)
						if err != nil {
							return nil, err
						}
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			}

			if !slices.Contains(s.Required, it.Key()) && !req.Context.Has(it.Key()) {
				fakes.Set(it.Key(), newFaker(func() (any, error) {
					n := gofakeit.Float32Range(0, 1)
					if n > 0.7 {
						return nil, nil
					}
					return f.fake()
				}))
			}
			fakes.Set(it.Key(), f)
		}
	}

	if s.PatternProperties != nil {
		for pattern, prop := range s.PatternProperties {
			if !strings.HasPrefix(pattern, "^") {
				pattern = "[a-zA-z0-9]*" + pattern
			}
			if !strings.HasSuffix(pattern, "$") {
				pattern = pattern + "[a-zA-z0-9]*"
			}
			re, err := syntax.Parse(pattern, syntax.Perl)
			if err != nil {
				return nil, fmt.Errorf("could not parse regex string: %v", pattern)
			}
			n := numPatterProperties()
			for i := 0; i < n; i++ {
				gen := regexGenerator{ra: req.g.rand}
				gen.regexGenerate(re, len(pattern)*100)
				propName := gen.sb.String()
				ex := propertyFromExample(propName, req)
				f, err := r.resolve(req.With(append(req.Path), prop, ex), fallback)
				if err != nil {
					return nil, err
				}
				fakes.Set(propName, f)
			}
		}
	}

	for _, name := range s.Required {
		if _, ok := fakes.Get(name); !ok {
			f, err := r.resolve(req.With(append(req.Path, name), nil, req.examples), false)
			if err != nil {
				return nil, err
			}
			fakes.Set(name, f)
		}
	}

	return fakes, nil
}

func (r *resolver) fakeDictionary(req *Request) (*sortedmap.LinkedHashMap[string, *faker], error) {
	length := numProperties(1, 10, req.Schema)
	fakes := &sortedmap.LinkedHashMap[string, *faker]{}
	for i := 0; i < length; i++ {
		f, err := r.resolve(req.WithSchema(req.Schema.AdditionalProperties), true)
		if err != nil {
			return nil, err
		}
		key := fakeDictionaryKey()
		fakes.Set(key, f)
	}
	return fakes, nil
}

func fakeDictionaryKey() string {
	key := gofakeit.Noun()
	return firstLetterToLower(key)
}

func firstLetterToLower(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])

	return string(r)
}

func numProperties(min, max int, s *schema.Schema) int {
	if s.MinProperties != nil {
		min = *s.MinProperties
	} else if s.Required != nil {
		min = len(s.Required)
	}
	if s.MaxProperties != nil {
		max = *s.MaxProperties
	} else if s.Required != nil {
		if len(s.Required) > max {
			max = len(s.Required)
		} else {
			n := gofakeit.Float32Range(0, 1)
			if n < 0.8 {
				max = len(s.Required)
			}
		}
	}
	if min == max {
		return min
	} else {
		return gofakeit.Number(min, max)
	}
}

func numPatterProperties() int {
	n, err := gofakeit.Weighted(numPatternProperties, weightsPatternProperties)
	if err != nil {
		return 1
	}
	return n.(int)
}

func isKnownDomain(r *Request) bool {
	if len(r.Path) == 0 {
		return false
	}
	domain := r.Path[len(r.Path)-1]

	for _, n := range g.root.Children {
		if n.Name == domain {
			return true
		}
		for _, attr := range g.root.Attributes {
			if attr == domain {
				return true
			}
		}
	}
	return false
}

func detectDomain(s *schema.Schema, root *Node) string {
	var attributes []string

	if s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			attributes = append(attributes, it.Key())
		}
	}

	if len(attributes) == 0 {
		return ""
	}

	var best *Node
	maxScore := 0.0

	for _, child := range root.Children {
		score := scoreDomain(attributes, child)
		if score > maxScore {
			maxScore = score
			best = child
		}
	}

	if best == nil {
		return ""
	}

	return best.Name
}

func scoreDomain(attribute []string, n *Node) float64 {
	score := 0.0
	for _, attr := range attribute {
		attr = strings.ToLower(attr)
		for _, child := range n.Children {
			if attr == child.Name {
				score += child.Weight
			}
		}
	}
	return score
}

func topologicalSort(fakes *sortedmap.LinkedHashMap[string, *faker]) ([]string, error) {
	inDegree := map[string]int{}
	graph := map[string][]string{}
	for it := fakes.Iter(); it.Next(); {
		inDegree[it.Key()] = 0
	}

	for it := fakes.Iter(); it.Next(); {
		key := it.Key()
		val := it.Value()
		if val.node == nil {
			continue
		}
		for _, dep := range val.node.DependsOn {
			if _, ok := inDegree[dep]; !ok {
				// Skip dependents that are not present
				continue
			}
			graph[dep] = append(graph[dep], key)
			inDegree[key]++
		}
	}

	// Queue of nodes with no dependencies
	var queue []string
	for it := fakes.Iter(); it.Next(); {
		if inDegree[it.Key()] == 0 {
			queue = append(queue, it.Key())
		}
	}

	var sorted []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(sorted) != fakes.Len() {
		return nil, fmt.Errorf("circular dependency detected")
	}
	return sorted, nil
}

func propertyFromExample(prop string, r *Request) []any {
	if r.examples == nil {
		return nil
	}

	var result []any
	for _, ex := range r.examples {
		if m, ok := ex.(map[string]interface{}); ok {
			result = append(result, m[prop])
		}
	}
	return result
}

func fakeByExample(r *Request) (*faker, error) {
	v, ok := example(r.Schema)
	if !ok {
		return nil, NoMatchFound
	}
	m := v.(map[string]any)
	f := func() (any, error) {
		return m, nil
	}
	return newFaker(f), nil
}

func examplesFromRequest(r *Request) []any {
	var result []any

	if r.examples != nil {
		result = append(result, r.examples...)
	}

	//mergeUnique(result, examples(r.Schema))
	result = append(result, examples(r.Schema)...)

	return result
}

func example(s *schema.Schema) (any, bool) {
	if s == nil || len(s.Examples) == 0 {
		return nil, false
	}

	index := gofakeit.Number(0, len(s.Examples)-1)
	return s.Examples[index].Value, true
}

func examples(s *schema.Schema) []any {
	if s == nil || len(s.Examples) == 0 {
		return nil
	}

	var result []any
	for _, e := range s.Examples {
		result = append(result, e.Value)
	}
	return result
}
