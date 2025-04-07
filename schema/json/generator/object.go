package generator

import (
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"unicode"
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

	switch {
	case !s.HasProperties() && s.AdditionalProperties != nil:
		fakes, err = r.fakeDictionary(req)
	case !s.HasProperties():
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
			f := fakes.Lookup(key)
			m[key], err = f.fake()
			if err != nil {
				return nil, err
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

	for it := s.Properties.Iter(); it.Next(); {
		prop := append(req.Path, it.Key())
		f, err := r.resolve(req.With(prop, it.Value()), fallback)
		if err != nil {
			if errors.Is(err, NoMatchFound) {
				if domain != "" {
					f, err = r.resolve(req.With([]string{domain, it.Key()}, it.Value()), true)
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
		fakes.Set(it.Key(), f)

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
	}
	if s.MaxProperties != nil {
		max = *s.MaxProperties
	}
	if min == max {
		return min
	} else {
		return gofakeit.Number(min, max)
	}
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
		for _, child := range n.Children {
			if attr == child.Name {
				score += child.Weight
			}
		}
	}
	return score
}

func fakeObject(r *Request) (interface{}, error) {
	s := r.Schema
	if s.Properties == nil {
		s.Properties = &schema.Schemas{LinkedHashMap: sortedmap.LinkedHashMap[string, *schema.Schema]{}}
		length := numProperties(0, 10, s)

		if length == 0 {
			return map[string]interface{}{}, nil
		}

		for i := 0; i < length; i++ {
			name := fakeDictionaryKey()
			s.Properties.Set(name, nil)
		}
	}

	m := map[string]any{}
	for it := s.Properties.Iter(); it.Next(); {
		v, err := New(r.With([]string{it.Key()}, it.Value()))
		if err != nil {
			return nil, err
		}
		m[it.Key()] = v
	}
	return m, nil
}

type objectFaker struct {
	key   string
	faker *faker
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

func fakeByExample(r *Request) (*faker, error) {
	index := gofakeit.Number(0, len(r.Schema.Examples)-1)
	v := r.Schema.Examples[index].Value.(map[string]any)
	f := func() (any, error) {
		return v, nil
	}
	return newFaker(f), nil
}
