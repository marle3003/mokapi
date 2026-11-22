package generator

import (
	"mokapi/schema/json/schema"
	"slices"
)

// shallowestLevels returns for each name the shallowest object level that declares a property
// with that name. The properties of the root object (or of the items of a root array) are
// level 1. Names that no schema declares are missing in the result.
//
// The resolver decides which parameter a property gets while it builds the fakers, depth-first,
// so a shallower sibling that is built later is not known yet. That is why the levels are
// computed in advance, level by level, following the keywords the resolver follows:
//   - properties and required names of an object, its children are one level deeper
//   - patternProperties and additionalProperties: only their schemas, one level deeper (the
//     names are random)
//   - items, prefixItems, contains: same level
//   - allOf, anyOf, oneOf, dependentSchemas, if/then/else: same level, the resolver builds
//     them from the request of the object they extend
//
// Accepted gaps: the resolver drops a property that hits the recursion guard, ignores
// properties that propertyNames rejects and subtracts sibling schemas of a oneOf. The
// levels are not adjusted for that.
func shallowestLevels(root *schema.Schema, names []string) map[string]int {
	levels := make(map[string]int)
	visited := newVisitedSchemas()
	current := []*schema.Schema{root}
	for level := 1; len(current) > 0 && len(levels) < len(names); level++ {
		var next []*schema.Schema
		// current grows while the schemas of the same level are expanded
		for i := 0; i < len(current); i++ {
			s := current[i]
			if s == nil || !visited.add(s) {
				continue
			}
			for _, name := range declaredNames(s) {
				if _, found := levels[name]; !found && slices.Contains(names, name) {
					levels[name] = level
				}
			}
			current = append(current, sameLevelSchemas(s)...)
			next = append(next, childSchemas(s)...)
		}
		current = next
	}
	return levels
}

func declaredNames(s *schema.Schema) []string {
	names := slices.Clone(s.Required)
	if s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			names = append(names, it.Key())
		}
	}
	return names
}

// sameLevelSchemas returns the schemas that describe the same object or its array items.
func sameLevelSchemas(s *schema.Schema) []*schema.Schema {
	result := []*schema.Schema{s.Items, s.Contains, s.If, s.Then, s.Else}
	result = append(result, s.PrefixItems...)
	result = append(result, s.AllOf...)
	result = append(result, s.AnyOf...)
	result = append(result, s.OneOf...)
	for _, ds := range s.DependentSchemas {
		result = append(result, ds)
	}
	return result
}

// childSchemas returns the schemas of the values of the object, one level deeper.
func childSchemas(s *schema.Schema) []*schema.Schema {
	var result []*schema.Schema
	if s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			result = append(result, it.Value())
		}
	}
	for _, p := range s.PatternProperties {
		result = append(result, p)
	}
	return append(result, s.AdditionalProperties)
}

// visitedSchemas identifies a schema like resolver.guardLoopLimit: by pointer or by reference.
type visitedSchemas struct {
	pointers map[*schema.Schema]struct{}
	refs     map[string]struct{}
}

func newVisitedSchemas() *visitedSchemas {
	return &visitedSchemas{pointers: map[*schema.Schema]struct{}{}, refs: map[string]struct{}{}}
}

// add returns false if the schema was already visited.
func (v *visitedSchemas) add(s *schema.Schema) bool {
	_, seen := v.pointers[s]
	if s.Ref != "" {
		_, seenRef := v.refs[s.Ref]
		seen = seen || seenRef
		v.refs[s.Ref] = struct{}{}
	}
	v.pointers[s] = struct{}{}
	return !seen
}
