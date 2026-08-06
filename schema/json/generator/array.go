package generator

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"reflect"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jinzhu/inflection"
)

func (r *resolver) resolveArray(req *Request) (*faker, error) {
	if f, err := findWithPlural(req); err == nil {
		return f, nil
	}

	s := req.Schema
	path := req.Path
	if len(path) > 0 {
		last := req.Path[len(req.Path)-1]
		singular := inflection.Singular(last)
		if singular != last {
			path = append(req.Path, singular)
		}
	}
	var item *faker
	var err error
	if s.Items != nil {
		if s.Items.Ref != "" {
			path = append(path, getPathFromRef(s.Items.Ref))
		}
		if len(s.Items.Enum) > 0 && (s.UniqueItems != nil && *s.UniqueItems) {
			index := gofakeit.Number(0, len(s.Items.Enum)-1)
			item = newFaker(func() (any, error) {
				index = (index + 1) % len(s.Items.Enum)
				return s.Items.Enum[index], nil
			})
		}
	}
	if item == nil {
		req.examples = examplesFromRequest(req)
		item, err = r.resolve(req.With(path, s.Items, itemsFromExample(req)), true)
	}
	if err != nil {
		var guard *RecursionGuard
		if errors.As(err, &guard) {
			if req.Schema.MinItems == nil || *req.Schema.MinItems == 0 {
				return newFaker(func() (any, error) {
					return []any{}, nil
				}), nil
			}
		}
		return nil, err
	}
	return newFaker(func() (interface{}, error) {
		return fakeArray(req, item)
	}), nil
}

func fakeArray(r *Request, fakeItem *faker) (interface{}, error) {
	s := r.Schema
	if s == nil {
		s = &schema.Schema{}
	}

	minItems := 0
	if s.MinItems != nil {
		minItems = *s.MinItems
	} else if s.MinContains != nil {
		minItems = *s.MinContains
	}

	// Determine enum size for items, if any
	enumSize := -1
	if s.Items != nil && len(s.Items.Enum) > 0 {
		enumSize = len(s.Items.Enum)
	}

	maxItemsWasSet := s.MaxItems != nil

	maxItems := minItems + 5
	if s.MaxItems != nil {
		maxItems = *s.MaxItems
	}

	// If maxItems wasn't explicitly set, don't let the default outrun the enum
	if !maxItemsWasSet && enumSize >= 0 && enumSize < maxItems {
		if enumSize > minItems {
			maxItems = enumSize
		} else {
			maxItems = minItems
		}
	}

	if maxItems < minItems {
		if s.MinItems != nil {
			return nil, errors.New("invalid schema: minItems must be less than maxItems")
		} else if s.MinContains != nil {
			return nil, errors.New("invalid schema: minContains must be less than maxItems")
		}
		return nil, fmt.Errorf("invalid schema: maxItems must be greater than minItems")
	}

	length := gofakeit.Number(minItems, maxItems)
	if s.Items != nil && s.Items.Boolean != nil && !*s.Items.Boolean {
		// disallows extra items in the tuple.
		length = 0
	}

	prefixItems := make([]any, 0, len(s.PrefixItems))
	containsMatches := 0
	for _, ps := range s.PrefixItems {
		r.Context.Snapshot()

		prefixItem := newFaker(func() (any, error) {
			return fakeBySchema(r.WithSchema(ps))
		})

		var v interface{}
		var err error
		if s.UniqueItems != nil && *s.UniqueItems {
			v, err = nextUnique(prefixItems, prefixItem.fake)
		} else {
			v, err = prefixItem.fake()
		}
		if err != nil {
			return nil, fmt.Errorf("%v: %v", err, s)
		}
		prefixItems = append(prefixItems, v)
		r.Context.Restore()

		if s.Contains != nil {
			p := parser.Parser{Schema: ps}
			if _, err = p.Parse(v); err == nil {
				containsMatches++
			}
		}
	}

	var containsFaker *faker
	containsNum := 0
	shuffleItems := s.ShuffleItems
	if s.Contains != nil {
		minContains := 1
		if s.MinContains != nil {
			minContains = *s.MinContains
		}
		maxContains := int(math.Max(float64(minContains), float64(length)))
		if s.MaxContains != nil {
			maxContains = *s.MaxContains
		}
		if err := validateContainsNum(minContains, maxContains); err != nil {
			return nil, err
		}
		containsFaker = newFaker(func() (any, error) {
			return fakeBySchema(r.WithSchema(s.Contains))
		})
		containsNum = gofakeit.Number(minContains, maxContains) - containsNum
		// Shuffle the array so the contains items are randomly distributed.
		shuffleItems = true
	}

	length = length - len(prefixItems)
	if length > 0 {
		arr := make([]any, 0, length)
		for i := 0; i < length; i++ {
			var nextItem *faker
			if containsFaker != nil && i < containsNum {
				nextItem = containsFaker
			} else {
				nextItem = fakeItem
			}

			err := fakeWithRetries(10, func() error {
				var v any
				var err error
				r.Context.Snapshot()
				defer r.Context.Restore()

				if s.UniqueItems != nil && *s.UniqueItems {
					v, err = nextUnique(arr, nextItem.fake)
				} else if enumSize > 0 {
					v, err = nextWithEnumBias(arr, enumSize, r.g.rand, nextItem.fake)
				} else {
					v, err = nextItem.fake()
				}

				if err != nil {
					return err
				}
				if s.Contains != nil && s.MaxContains != nil {
					p := parser.Parser{Schema: s.Contains}
					if _, errContains := p.Parse(v); errContains == nil {
						if containsMatches+1 > *s.MaxContains {

							return fmt.Errorf("reached maximum of value maxContains=%d", *s.MaxContains)
						} else {
							containsMatches++
						}
					}
				}

				arr = append(arr, v)
				return nil
			})
			if err != nil {
				if minItems < length {
					length--
					i--
				} else {
					return nil, fmt.Errorf("failed to generate valid array: %w", err)
				}
			}
		}

		if shuffleItems {
			r.g.rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
		}

		result := append(prefixItems, arr...)
		return result, nil
	} else {
		return prefixItems, nil
	}
}

// nextWithEnumBias softly discourages duplicate items when the item schema
// is an enum, without enforcing uniqueItems semantics. pUnique is derived
// from the fraction of the enum's distinct values already used, eased by
// an exponent so small enums don't collapse to near-random behavior after
// just one or two picks. Duplicates remain legal; this only shapes how
// often we bother trying to avoid them.
const enumBiasExponent = 3.0

func nextWithEnumBias(arr []any, enumSize int, rnd *rand.Rand, fake func() (any, error)) (any, error) {
	distinctUsed := countDistinct(arr)
	x := float64(distinctUsed) / float64(enumSize)
	pUnique := 1 - math.Pow(x, enumBiasExponent)
	if pUnique < 0 {
		pUnique = 0
	}

	if rnd.Float64() < pUnique {
		// A few short retries to land on a non-duplicate; fall back to a
		// plain draw rather than erroring, since duplicates are legal here.
		for i := 0; i < 5; i++ {
			v, err := fake()
			if err != nil {
				return nil, err
			}
			if !contains(arr, v) {
				return v, nil
			}
		}
	}

	return fake()
}

func nextUnique(arr []interface{}, fakeItem func() (interface{}, error)) (interface{}, error) {
	for i := 0; i < 10; i++ {
		v, err := fakeItem()
		if err != nil {
			return nil, err
		}
		if !contains(arr, v) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("cannot fill array with unique items")
}

func contains(s []interface{}, v interface{}) bool {
	for _, i := range s {
		if reflect.DeepEqual(i, v) {
			return true
		}
	}
	return false
}

func itemsFromExample(r *Request) []any {
	var result []any
	for _, e := range r.examples {
		if arr, ok := e.([]any); ok {
			result = append(result, arr...)
		}
	}
	return result
}

func findWithPlural(req *Request) (*faker, error) {
	if len(req.Path) == 0 {
		return nil, NotSupported
	}

	last := req.Path[len(req.Path)-1]
	plural := inflection.Plural(last)
	if plural != last {
		req.Path = append(req.Path[:len(req.Path)-1], plural)
	}

	path := tokenize(req.Path)
	n := findBestMatch(g.root, req.WithPath(path))
	if n != nil && !n.isRootOrDefault() && n.Fake != nil {
		return newFakerWithFallback(n, req), nil
	}
	return nil, NotSupported
}

func validateContainsNum(min, max int) error {
	if min > max {
		return fmt.Errorf("invalid minContains '%v' and maxContains '%v'", min, max)
	}
	if min < 0 {
		return fmt.Errorf("invalid minContains '%v': must be a non-negative number", min)
	}
	if max < 0 {
		return fmt.Errorf("invalid maxContains '%v': must be a non-negative number", min)
	}
	return nil
}

func countDistinct(arr []any) int {
	seen := make(map[string]struct{}, len(arr))
	for _, v := range arr {
		seen[fmt.Sprintf("%v", v)] = struct{}{}
	}
	return len(seen)
}
