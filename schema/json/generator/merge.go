package generator

import (
	"fmt"
	"math/big"
	"mokapi/schema/json/schema"
	"slices"
)

// given any schemas, return one schema with all constraints applied.
func intersectSchemas(schemas ...*schema.Schema) (*schema.Schema, error) {
	result := &schema.Schema{}
	var err error
	for _, s := range schemas {
		result, err = intersectSchema(result, s)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// given two schemas, return one schema with both constraints applied.
func intersectSchema(base, s *schema.Schema) (*schema.Schema, error) {
	if base == nil {
		return s, nil
	}
	if s == nil {
		return base, nil
	}

	result := base.Clone()
	// additionalProperties only recognizes properties declared in the same subschema as itself,
	// so we reset this
	result.AdditionalProperties = nil

	var err error

	if len(base.Type) > 0 && len(s.Type) > 0 {
		result.Type = nil
		for _, t := range s.Type {
			if slices.Contains(base.Type, t) {
				result.Type = append(result.Type, t)
			}
		}
		if len(result.Type) == 0 {
			return nil, fmt.Errorf("no shared types found: %s and %s", base.Type, s.Type)
		}
	} else if len(s.Type) > 0 {
		result.Type = s.Type
	}

	result.Enum = append(result.Enum, s.Enum...)
	if s.Const != nil {
		result.Const = s.Const
	}
	if s.MultipleOf != nil {
		if result.MultipleOf != nil {
			a := big.Rat{}
			a.SetFloat64(*result.MultipleOf)
			b := big.Rat{}
			b.SetFloat64(*s.MultipleOf)
			n := lcmMultipleOf(&a, &b)
			f, _ := n.Float64()
			result.MultipleOf = &f
		} else {
			result.MultipleOf = s.MultipleOf
		}
	}
	if s.Maximum != nil {
		if result.Maximum == nil || *s.Maximum < *result.Maximum {
			result.Maximum = s.Maximum
		}
	}
	if s.ExclusiveMaximum != nil {
		if result.ExclusiveMaximum != nil {
			if s.ExclusiveMaximum.IsA() {
				if !result.ExclusiveMaximum.IsA() || s.ExclusiveMaximum.A < result.ExclusiveMaximum.A {
					result.ExclusiveMaximum.A = s.ExclusiveMaximum.A
				}
			} else if s.ExclusiveMaximum.B && !result.ExclusiveMaximum.IsA() {
				result.ExclusiveMaximum = s.ExclusiveMaximum
			}
		} else {
			result.ExclusiveMaximum = s.ExclusiveMaximum
		}
	}
	if s.Minimum != nil {
		if result.Minimum == nil || *s.Minimum > *result.Minimum {
			result.Minimum = s.Minimum
		}
	}
	if s.ExclusiveMinimum != nil {
		if result.ExclusiveMinimum != nil {
			if s.ExclusiveMinimum.IsA() {
				if !result.ExclusiveMinimum.IsA() || s.ExclusiveMinimum.A > result.ExclusiveMinimum.A {
					result.ExclusiveMaximum.A = s.ExclusiveMaximum.A
				}
			} else if s.ExclusiveMinimum.B && !result.ExclusiveMinimum.IsA() {
				result.ExclusiveMinimum = s.ExclusiveMinimum
			}
		} else {
			result.ExclusiveMinimum = s.ExclusiveMinimum
		}
	}

	result.MaxLength = mergeMax(result.MaxLength, s.MaxLength)
	result.MinLength = mergeMin(result.MinLength, s.MinLength)

	if s.Pattern != "" && result.Pattern == "" {
		result.Pattern = s.Pattern
	}
	if s.Format != "" && result.Format == "" {
		result.Format = s.Format
	}
	if s.Items != nil {
		if result.Items == nil {
			result.Items = s.Items
		} else {
			result.Items, err = intersectSchema(result.Items, s.Items)
			if err != nil {
				return nil, err
			}
		}
	}
	if s.PrefixItems != nil {
		result.PrefixItems = append(result.PrefixItems, s.PrefixItems...)
	}
	if s.UnevaluatedItems != nil {
		if result.UnevaluatedItems == nil {
			result.UnevaluatedItems = s.UnevaluatedItems
		} else {
			result.UnevaluatedItems, err = intersectSchema(result.UnevaluatedItems, s.UnevaluatedItems)
			if err != nil {
				return nil, err
			}
		}
	}
	if s.Contains != nil {
		if result.Contains == nil {
			result.Contains = s.Contains
		} else {
			result.Contains, err = intersectSchema(result.Contains, s.Contains)
			if err != nil {
				return nil, err
			}
		}
	}

	result.MaxContains = mergeMax(result.MaxContains, s.MaxContains)
	result.MinContains = mergeMin(result.MinContains, s.MinContains)

	result.MaxItems = mergeMax(result.MaxItems, s.MaxItems)
	result.MinItems = mergeMin(result.MinItems, s.MinItems)

	if s.UniqueItems != nil && *s.UniqueItems {
		result.UniqueItems = s.UniqueItems
	}

	if s.ShuffleItems {
		result.ShuffleItems = s.ShuffleItems
	}

	if s.Properties != nil {
		if result.Properties == nil {
			result.Properties = s.Properties
		} else {
			for it := s.Properties.Iter(); it.Next(); {
				resultVal := result.Properties.Get(it.Key())
				merged, err := intersectSchema(resultVal, it.Value())
				if err != nil {
					return nil, err
				}
				result.Properties.Set(it.Key(), merged)
			}
		}
	}

	if s.PatternProperties != nil {
		if result.PatternProperties == nil {
			result.PatternProperties = s.PatternProperties
		} else {
			for k, v := range s.PatternProperties {
				result.PatternProperties[k], err = intersectSchema(result.PatternProperties[k], v)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	result.MaxProperties = mergeMax(result.MaxProperties, s.MaxProperties)
	result.MinProperties = mergeMin(result.MinProperties, s.MinProperties)

	for _, req := range s.Required {
		if !slices.Contains(result.Required, req) {
			result.Required = append(result.Required, req)
		}
	}
	return result, nil
}

func mergeMin(a, b *int) *int {
	if b != nil {
		if a == nil || *b > *a {
			return b
		}
	}
	return a
}

func mergeMax(a, b *int) *int {
	if b != nil {
		if a == nil || *b > *a {
			return b
		}
	}
	return a
}

// gcd computes greatest common divisor of a and b
func gcd(a, b *big.Int) *big.Int {
	return new(big.Int).GCD(nil, nil, a, b)
}

// lcm computes least common multiple of a and b
func lcm(a, b *big.Int) *big.Int {
	if a.Sign() == 0 || b.Sign() == 0 {
		return big.NewInt(0)
	}
	g := gcd(a, b)
	return new(big.Int).Div(new(big.Int).Mul(a, b), g)
}

// lcmMultipleOf returns the combined multipleOf constraint of two values
func lcmMultipleOf(a, b *big.Rat) *big.Rat {
	// a = an/ad, b = bn/bd
	an, ad := a.Num(), a.Denom()
	bn, bd := b.Num(), b.Denom()

	// bring to common denominator
	// a = an/ad, b = bn/bd
	// combined denominator = lcm(ad, bd)
	den := lcm(ad, bd)

	// scale numerators
	scaleA := new(big.Int).Div(den, ad)
	scaleB := new(big.Int).Div(den, bd)
	anScaled := new(big.Int).Mul(an, scaleA)
	bnScaled := new(big.Int).Mul(bn, scaleB)

	// take lcm of numerators
	num := lcm(anScaled, bnScaled)

	// result = num/den
	return new(big.Rat).SetFrac(num, den)
}
