package schema

import (
	"fmt"
	"strings"
)

func (s *Schema) String() string {
	var sb strings.Builder

	if len(s.AnyOf) > 0 {
		sb.WriteString("any of ")
		for _, i := range s.AnyOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(i.String())
		}
		return sb.String()
	}
	if len(s.AllOf) > 0 {
		sb.WriteString("all of ")
		for _, r := range s.AllOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(r.String())
		}
		return sb.String()
	}
	if len(s.OneOf) > 0 {
		sb.WriteString("one of ")
		for _, r := range s.OneOf {
			if sb.Len() > 7 {
				sb.WriteString(", ")
			}
			sb.WriteString(r.String())
		}
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("schema type=%v", s.Type.String()))

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

	if s.IsObject() && s.Properties != nil {
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
		sb.WriteString(" items=")
		sb.WriteString(fmt.Sprintf("%v", s.Items.Value))
	}

	return sb.String()
}

func (t *Types) String() string {
	if len(*t) == 1 {
		return (*t)[0]
	} else if len(*t) > 1 {
		return fmt.Sprintf("%v", *t)
	}
	return ""
}

//func (s *Schema) IsFreeForm() bool {
//	if s.Type != "object" {
//		return false
//	}
//	free := s.Type == "object" && (s.Properties == nil || s.Properties.Len() == 0)
//	if s.AdditionalProperties == nil || free {
//		return true
//	}
//	return s.AdditionalProperties.IsFreeForm()
//}
//
//func (s *Schema) IsDictionary() bool {
//	return s.AdditionalProperties != nil && s.AdditionalProperties.Ref != nil && s.AdditionalProperties.Value != nil && s.AdditionalProperties.Value.Type != ""
//}
