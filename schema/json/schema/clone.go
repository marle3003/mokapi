package schema

func (s *Schema) Clone() *Schema {
	if s == nil {
		return nil
	}

	clone := &Schema{
		Id:            s.Id,
		Ref:           s.Ref,
		DynamicRef:    s.DynamicRef,
		Schema:        s.Schema,
		Boolean:       s.Boolean,
		Anchor:        s.Anchor,
		DynamicAnchor: s.DynamicAnchor,

		Const: s.Const,

		MultipleOf: s.MultipleOf,
		Maximum:    s.Maximum,
		Minimum:    s.Minimum,

		MaxLength: s.MaxLength,
		MinLength: s.MinLength,
		Pattern:   s.Pattern,
		Format:    s.Format,

		Items:            s.Items.Clone(),
		UnevaluatedItems: s.UnevaluatedItems.Clone(),
		Contains:         s.Contains.Clone(),
		MaxContains:      s.MaxContains,
		MinContains:      s.MinContains,
		MaxItems:         s.MaxItems,
		MinItems:         s.MinItems,
		UniqueItems:      s.UniqueItems,
		ShuffleItems:     s.ShuffleItems,

		MaxProperties:         s.MaxProperties,
		MinProperties:         s.MinProperties,
		AdditionalProperties:  s.AdditionalProperties.Clone(),
		UnevaluatedProperties: s.UnevaluatedProperties.Clone(),
		PropertyNames:         s.PropertyNames.Clone(),

		Not: s.Not.Clone(),

		If:   s.If.Clone(),
		Then: s.Then.Clone(),
		Else: s.Else.Clone(),

		Title:       s.Title,
		Description: s.Description,
		Default:     s.Default,
		Deprecated:  s.Deprecated,
		Examples:    s.Examples,

		ContentMediaType: s.ContentMediaType,
		ContentEncoding:  s.ContentEncoding,
	}

	for _, t := range s.Type {
		clone.Type = append(clone.Type, t)
	}

	for _, e := range s.Enum {
		clone.Enum = append(clone.Enum, e)
	}

	if s.ExclusiveMaximum != nil {
		clone.ExclusiveMaximum = &UnionType[float64, bool]{
			A: s.ExclusiveMaximum.A,
			B: s.ExclusiveMaximum.B,
		}
	}
	if s.ExclusiveMinimum != nil {
		clone.ExclusiveMinimum = &UnionType[float64, bool]{
			A: s.ExclusiveMinimum.A,
			B: s.ExclusiveMinimum.B,
		}
	}

	for _, pi := range s.PrefixItems {
		clone.PrefixItems = append(clone.PrefixItems, pi.Clone())
	}

	if s.Properties != nil {
		clone.Properties = new(Schemas)
		for it := s.Properties.Iter(); it.Next(); {
			clone.Properties.Set(it.Key(), it.Value().Clone())
		}
	}

	if s.PatternProperties != nil {
		clone.PatternProperties = map[string]*Schema{}
		for k, v := range s.PatternProperties {
			clone.PatternProperties[k] = v.Clone()
		}
	}

	if s.DependentRequired != nil {
		clone.DependentRequired = map[string][]string{}
		for k, v := range s.DependentRequired {
			clone.DependentRequired[k] = append(clone.DependentRequired[k], v...)
		}
	}

	if s.DependentSchemas != nil {
		clone.DependentSchemas = map[string]*Schema{}
		for k, v := range s.DependentSchemas {
			clone.DependentSchemas[k] = v.Clone()
		}
	}

	for _, v := range s.Required {
		clone.Required = append(clone.Required, v)
	}

	for _, as := range s.AllOf {
		clone.AllOf = append(clone.AllOf, as.Clone())
	}
	for _, as := range s.AnyOf {
		clone.AnyOf = append(clone.AnyOf, as.Clone())
	}
	for _, as := range s.OneOf {
		clone.OneOf = append(clone.OneOf, as.Clone())
	}

	return clone
}
