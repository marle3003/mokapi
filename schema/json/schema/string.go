package schema

import (
	"fmt"
	"strings"
)

func (s *Schema) String() string {
	var sb strings.Builder

	if s.Boolean != nil {
		if *s.Boolean {
			return "schema (always valid)"
		}
		return "schema (always invalid)"
	}

	if len(s.AnyOf) > 0 {
		sb.WriteString("any of ")
		sb.WriteString(subSchemaString(s.AnyOf))
		return sb.String()
	}
	if len(s.AllOf) > 0 {
		sb.WriteString("all of ")
		sb.WriteString(subSchemaString(s.AllOf))
		return sb.String()
	}
	if len(s.OneOf) > 0 {
		sb.WriteString("one of ")
		sb.WriteString(subSchemaString(s.OneOf))
		return sb.String()
	}

	if len(s.Type) > 0 {
		sb.WriteString(fmt.Sprintf("type=%v", s.Type.String()))
	}

	if len(s.Format) > 0 {
		sb.WriteString(fmt.Sprintf(" format=%v", s.Format))
	}
	if len(s.Pattern) > 0 {
		sb.WriteString(fmt.Sprintf(" pattern=%v", s.Pattern))
	}
	if s.MinLength != nil {
		sb.WriteString(fmt.Sprintf(" minLength=%v", *s.MinLength))
	}
	if s.MaxLength != nil {
		sb.WriteString(fmt.Sprintf(" maxLength=%v", *s.MaxLength))
	}
	if s.Minimum != nil {
		sb.WriteString(fmt.Sprintf(" minimum=%v", *s.Minimum))
	}
	if s.Maximum != nil {
		sb.WriteString(fmt.Sprintf(" maximum=%v", *s.Maximum))
	}
	if s.MultipleOf != nil {
		sb.WriteString(fmt.Sprintf(" multipleOf=%v", *s.MultipleOf))
	}
	if s.ExclusiveMinimum != nil {
		if s.ExclusiveMinimum.IsA() {
			sb.WriteString(fmt.Sprintf(" exclusiveMinimum=%v", s.ExclusiveMinimum.A))
		} else {
			sb.WriteString(fmt.Sprintf(" exclusiveMinimum=%v", s.ExclusiveMinimum.B))
		}
	}
	if s.ExclusiveMaximum != nil {
		if s.ExclusiveMaximum.IsA() {
			sb.WriteString(fmt.Sprintf(" exclusiveMaximum=%v", s.ExclusiveMaximum.A))
		} else {
			sb.WriteString(fmt.Sprintf(" exclusiveMaximum=%v", s.ExclusiveMaximum.B))
		}
	}
	if s.MinItems != nil {
		sb.WriteString(fmt.Sprintf(" minItems=%v", *s.MinItems))
	}
	if s.MaxItems != nil {
		sb.WriteString(fmt.Sprintf(" maxItems=%v", *s.MaxItems))
	}
	if s.MinProperties != nil {
		sb.WriteString(fmt.Sprintf(" minProperties=%v", *s.MinProperties))
	}
	if s.MaxProperties != nil {
		sb.WriteString(fmt.Sprintf(" maxProperties=%v", *s.MaxProperties))
	}
	if s.UniqueItems {
		sb.WriteString(" unique-items")
	}

	if s.Properties != nil {
		var sbProp strings.Builder
		for _, p := range s.Properties.Keys() {
			if sbProp.Len() > 0 {
				sbProp.WriteString(", ")
			}
			sbProp.WriteString(fmt.Sprintf("%v", p))
		}
		sb.WriteString(fmt.Sprintf(" properties=[%v]", sbProp.String()))
	}
	if len(s.Required) > 0 {
		sb.WriteString(fmt.Sprintf(" required=%v", s.Required))
	}
	if s.Type.Includes("object") && !s.IsFreeForm() {
		sb.WriteString(" free-form=false")
	}

	if s.IsArray() && s.Items != nil {
		sb.WriteString(fmt.Sprintf(" items=(%v)", s.Items))
	}

	if s.Not != nil {
		sb.WriteString(fmt.Sprintf(" not (%v)", s.Not))
	}

	if len(s.Title) > 0 {
		sb.WriteString(fmt.Sprintf(" title=%v", s.Title))
	} else if len(s.Description) > 0 {
		sb.WriteString(fmt.Sprintf(" description=%v", s.Description))
	}

	if sb.Len() == 0 {
		return "empty schema"
	}

	str := sb.String()
	if string(str[0]) == " " {
		str = str[1:]
	}

	return "schema " + str
}

func subSchemaString(subs []*Schema) string {
	var sb strings.Builder
	sb.WriteString("(")
	for i, sub := range subs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sub.String())
	}
	sb.WriteString(")")
	return sb.String()
}
