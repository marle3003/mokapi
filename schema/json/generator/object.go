package generator

import (
	"errors"
	"fmt"
	"maps"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"regexp/syntax"
	"slices"
	"sort"
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

	if !s.HasProperties() && s.PatternProperties == nil && s.AdditionalProperties == nil {
		if len(s.Examples) > 0 {
			return fakeByExample(req)
		}
		match := findBestMatch(g.root, req)
		return newFakerWithFallback(match, req), nil
	}

	var fakes *sortedmap.LinkedHashMap[string, *faker]
	var err error

	fakes, err = r.fakeObject(req)

	if err != nil {
		return nil, err
	}

	fake := func() (interface{}, error) {
		var propNames []string
		props := map[string]interface{}{}
		var sorted []string
		sorted, err = topologicalSort(fakes)
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		p := parser.Parser{Schema: s, ValidateAdditionalProperties: true}
		err = fakeWithRetries(10, func() error {
			for _, key := range sorted {
				f := fakes.Lookup(key)
				props[key], err = f.fake()
				if !slices.Contains(s.Required, key) {
					propNames = append(propNames, key)
				}
				if err != nil {
					return err
				}
			}

			result = map[string]interface{}{}
			for _, key := range s.Required {
				result[key] = props[key]
			}

			changed := true
			attempts := 0
			for changed {
				changed = false
				if attempts >= 10 {
					return fmt.Errorf("cannot satisfy conditions")
				}
				attempts++

				minProps := 0
				if s.MinProperties != nil {
					minProps = *s.MinProperties
					if minProps-len(s.Required) < 0 {
						return fmt.Errorf("invalid schema: minProperties must be at least the number of required properties")
					}
				}
				maxProps := -1
				if s.MaxProperties != nil {
					maxProps = *s.MaxProperties
					if maxProps-len(s.Required) < 0 {
						return fmt.Errorf("invalid schema: maxProperties must be at least the number of required properties")
					}
				}

				// using array to loop to get predictable result for tests
				// shuffle propNames to get random optional properties
				req.g.rand.Shuffle(len(propNames), func(i, j int) { propNames[i], propNames[j] = propNames[j], propNames[i] })
				for _, k := range propNames {
					if _, ok := result[k]; ok {
						continue
					}

					n := len(result)
					if n >= minProps {
						n := gofakeit.Float64Range(0, 1)
						if n > req.g.cfg.OptionalPropertiesProbability() {
							continue
						}
					}
					if maxProps >= 0 && n >= maxProps {
						break
					}
					result[k] = props[k]
				}

				// apply if-then-else
				err = applyConditional(req, result, &changed)
				if err != nil {
					return err
				}

				// apply dependentRequired
				for name, list := range s.DependentRequired {
					if _, ok := result[name]; ok {
						newLength := len(result) + len(list)
						if s.MaxProperties == nil || newLength <= *s.MaxProperties {
							for _, required := range list {
								if _, ok = result[required]; !ok {
									result[required] = props[required]
									changed = true
								}
							}
						} else if !slices.Contains(s.Required, name) {
							delete(result, name)
							changed = true
						} else {
							return fmt.Errorf("cannot apply dependentRequired for '%s': maxProperties=%d was exceeded", name, *s.MaxProperties)
						}
					}
				}

				// apply dependentSchemas
				for name, ds := range s.DependentSchemas {
					if _, ok := result[name]; ok {
						var f *faker
						f, err = r.resolve(req.WithSchema(ds), true)
						if err != nil {
							return err
						}
						var v any
						v, err = f.fake()
						if err != nil {
							return err
						}
						if m, ok := v.(map[string]any); ok {
							newLength := len(result) + len(m)
							if s.MaxProperties == nil || newLength <= *s.MaxProperties {
								for propKey, propValue := range m {
									if _, ok = result[propKey]; !ok {
										result[propKey] = propValue
										changed = true
									}
								}
								return nil
							} else if !slices.Contains(s.Required, name) {
								delete(result, name)
								changed = true
							} else {
								return fmt.Errorf("cannot apply dependentSchemas for '%s': maxProperties=%d was exceeded", name, *s.MaxProperties)
							}
						}
					}
				}

				err = applyObjectAnyOf(req, result, &changed)
				if err != nil {
					return err
				}
				err = applyObjectAllOf(req, result, &changed)
				if err != nil {
					return err
				}
				err = applyOneOf(req, result, &changed)
				if err != nil {
					return err
				}
			}

			_, err = p.Parse(result)

			return err
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate valid object: %w", err)
		}

		return result, nil
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
	propertyNameParser := propertyNamesParser(s)

	if s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			propName := it.Key()
			if _, err := propertyNameParser.Parse(propName); err != nil {
				continue
			}
			prop := append(req.Path, propName)
			ex := propertyFromExample(propName, req)
			f, err := r.resolve(req.With(prop, it.Value(), ex), fallback)
			if err != nil {
				var guard *RecursionGuard
				if errors.As(err, &guard) {
					if !slices.Contains(req.Schema.Required, propName) {
						continue
					}
				}
				if errors.Is(err, NoMatchFound) {
					if domain != "" {
						f, err = r.resolve(req.With([]string{domain, propName}, it.Value(), ex), true)
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

			fakes.Set(propName, f)
		}
	}

	if s.PatternProperties != nil {
		// Collect and sort keys to get a fixed order of iteration
		keys := make([]string, 0, len(s.PatternProperties))
		for k := range s.PatternProperties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, pattern := range keys {
			prop := s.PatternProperties[pattern]
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
			fmt.Printf("numPatterProperties: %v\n", n)
			for i := 0; i < n; i++ {
				gen := regexGenerator{ra: req.g.rand}
				gen.regexGenerate(re, len(pattern)*100)
				propName := gen.sb.String()
				ex := propertyFromExample(propName, req)
				f, err := r.resolve(req.With(append(req.Path), prop, ex), fallback)
				if err != nil {
					return nil, err
				}
				if _, err = propertyNameParser.Parse(propName); err != nil {
					continue
				}
				fakes.Set(propName, f)
			}
		}
	}

	if s.AdditionalProperties != nil && s.AdditionalProperties.Boolean == nil {
		// if additionalProperties=false no additional properties is allowed
		// if additionalProperties=true we don't add random properties, it is not expected by users

		length := numProperties(1, 10, req.Schema)
		for i := 0; i < length; i++ {
			f, err := r.resolve(req.WithSchema(req.Schema.AdditionalProperties), true)
			if err != nil {
				return nil, err
			}
			key, err := newPropertyName(propertyNameParser)
			if err != nil {
				continue
			}
			fakes.Set(key, f)
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
	propertyNameParser := propertyNamesParser(req.Schema)
	for i := 0; i < length; i++ {
		f, err := r.resolve(req.WithSchema(req.Schema.AdditionalProperties), true)
		if err != nil {
			return nil, err
		}
		key, err := newPropertyName(propertyNameParser)
		if err != nil {
			continue
		}
		fakes.Set(key, f)
	}
	return fakes, nil
}

func newPropertyName(propertyNameParser *parser.Parser) (string, error) {
	key := gofakeit.Noun()
	if _, err := propertyNameParser.Parse(key); err != nil {
		var v any
		v, err = New(NewRequest(nil, propertyNameParser.Schema, nil))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", v), nil
	}
	return firstLetterToLower(key), nil
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
	m, ok := v.(map[string]any)
	if !ok {
		return nil, NoMatchFound
	}
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

func propertyNamesParser(s *schema.Schema) *parser.Parser {
	p := &parser.Parser{}
	if s == nil || s.PropertyNames == nil {
		p.Schema = &schema.Schema{Type: schema.Types{"string"}}
	} else {
		p.Schema = s.PropertyNames
	}
	return p
}

func applyConditional(req *Request, result map[string]any, changed *bool) error {
	if req.Schema.If == nil {
		return nil
	}
	s := req.Schema
	var err error

	numOfRemovableProperties := len(result) - len(s.Required)
	p := parser.Parser{}
	conditionApplied := false
	for ; numOfRemovableProperties >= 0; numOfRemovableProperties-- {
		err = fakeWithRetries(10, func() error {
			var condResult map[string]any

			_, err = p.ParseWith(result, s.If)
			var cond *schema.Schema
			if err == nil && s.Then != nil {
				cond = s.Then
			} else if err != nil && s.Else != nil {
				cond = s.Else
			}
			if cond != nil {
				var v any
				v, err = New(req.WithSchema(cond))
				if err != nil {
					return err
				}
				if m, ok := v.(map[string]any); ok {
					condResult = m
				} else {
					return fmt.Errorf("invalid conditional schema: got %s, expected Object", cond.Type)
				}
			}

			newLength := len(result) + len(condResult)
			if s.MaxProperties == nil || newLength <= *s.MaxProperties {
				for k, v := range condResult {
					result[k] = v
				}
				return nil
			}
			return fmt.Errorf("reached maximum of value maxProperties=%d", *s.MaxProperties)
		})

		if err == nil {
			conditionApplied = true
			break
		}

		// remove one optional property to try conditional again
		for k := range result {
			if slices.Contains(s.Required, k) {
				continue
			}
			delete(result, k)
			*changed = true
			break
		}
	}
	if !conditionApplied {
		return fmt.Errorf("conditional schema could not be applied: %w", err)
	}
	return nil
}

func applyObjectAnyOf(req *Request, result map[string]any, changed *bool) error {
	base := req.Schema
	if base.AnyOf == nil || len(base.AnyOf) == 0 {
		return nil
	}

	var err error
	isOneValid := false
	p := parser.Parser{}
	for _, as := range base.AnyOf {
		p.Schema = as
		if _, err = p.Parse(result); err == nil {
			isOneValid = true
			break
		}
	}
	if !isOneValid {
		err = fakeWithRetries(10, func() error {
			i := gofakeit.Number(0, len(base.AnyOf)-1)
			as, err := extendBranchWithBase(base.AnyOf[i], base)
			if err != nil {
				return fmt.Errorf("cannot extend anyOf: %w", err)
			}
			var resultAny any
			resultAny, err = New(req.WithSchema(as))
			if m, ok := resultAny.(map[string]any); ok {

				newLength := len(result) + len(m)
				if base.MaxProperties == nil || newLength <= *base.MaxProperties {
					for k, v := range m {
						result[k] = v
					}
					*changed = true
					return nil
				}
				return fmt.Errorf("reached maximum of value maxProperties=%d", *base.MaxProperties)

			} else {
				return fmt.Errorf("invalid conditional schema: got %s, expected Object", as.Type)
			}
		})
		if err != nil {
			return fmt.Errorf("cannot apply one schema of 'anyOf': %w", err)
		}
	}

	return nil
}

func applyObjectAllOf(req *Request, result map[string]any, changed *bool) error {
	base := req.Schema
	if base.AllOf == nil || len(base.AllOf) == 0 {
		return nil
	}

	intersection, err := intersectSchemas(base.AllOf...)
	if err != nil {
		return err
	}

	p := parser.Parser{Schema: intersection}
	if _, err = p.Parse(result); err == nil {
		return nil
	}
	as, err := extendBranchWithBase(intersection, base)
	if err != nil {
		return fmt.Errorf("cannot extend allOf: %w", err)
	}

	err = fakeWithRetries(10, func() error {
		var resultAll any
		resultAll, err = New(req.WithSchema(as))
		if m, ok := resultAll.(map[string]any); ok {

			newLength := len(result) + len(m)
			if base.MaxProperties == nil || newLength <= *base.MaxProperties {
				for k, v := range m {
					result[k] = v
				}
				*changed = true
				return nil
			}
			return fmt.Errorf("reached maximum of value maxProperties=%d", *base.MaxProperties)

		} else {
			return fmt.Errorf("invalid conditional schema: got %s, expected Object", as.Type)
		}
	})
	if err != nil {
		return fmt.Errorf("cannot apply one schema of 'allOf': %w", err)
	}

	return nil
}

func applyOneOf(req *Request, result map[string]any, changed *bool) error {
	base := req.Schema
	if base.OneOf == nil || len(base.OneOf) == 0 {
		return nil
	}
	p := parser.Parser{Schema: base, ValidateAdditionalProperties: true}
	if _, err := p.Parse(result); err == nil {
		return nil
	}

	index := gofakeit.Number(0, len(base.OneOf)-1)
	var err error
	for i := 0; i < len(base.OneOf); i++ {

		selected := selectIndexAndSubtractOthers(index, base.OneOf...)
		selected, err = extendBranchWithBase(selected, base)

		err = fakeWithRetries(10, func() error {
			var resultOne any
			resultOne, err = New(req.WithSchema(selected))
			if m, ok := resultOne.(map[string]any); ok {
				var temp = maps.Clone(result)
				for k, v := range m {
					temp[k] = v
				}

				for idx, one := range base.OneOf {
					if idx == index {
						continue
					}
					_, err = p.ParseWith(temp, one)
					if err == nil {
						return fmt.Errorf("data is valid against more of the given oneOf subschemas")
					}
				}

				newLength := len(result) + len(m)
				if base.MaxProperties == nil || newLength <= *base.MaxProperties {
					for k, v := range m {
						result[k] = v
					}
					*changed = true
					return nil
				}
				return fmt.Errorf("reached maximum of value maxProperties=%d", *base.MaxProperties)

			} else {
				return fmt.Errorf("invalid conditional schema: got %s, expected Object", selected.Type)
			}
		})
		if err == nil {
			return nil
		}
		index = (index + 1) % len(base.OneOf)

	}
	if err != nil {
		return fmt.Errorf("cannot apply one of the subschemas in 'oneOf': %w", err)
	}

	return nil
}
