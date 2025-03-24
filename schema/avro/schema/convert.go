package schema

import (
	json "mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func (s *Schema) Convert() *json.Schema {
	js := &json.Schema{
		Title:       s.fullname,
		Description: s.Doc,
	}

	if len(s.Type) > 1 {
		for _, t := range s.Type {
			switch v := t.(type) {
			case string:
				name := getFullname(s, v)
				if named, ok := table[name]; ok {
					js.AnyOf = append(js.AnyOf, named.Convert())
				} else {
					jsAny := &json.Schema{Type: json.Types{getJsonType(v)}}
					if v == "int" {
						jsAny.Format = "int32"
					}
					js.AnyOf = append(js.AnyOf, jsAny)
				}
			case *Schema:
				js.AnyOf = append(js.AnyOf, v.Convert())
			}
		}
		return js
	}
	if len(s.Type) == 1 {
		if str, ok := s.Type[0].(string); ok {
			name := getFullname(s, str)
			if named, ok := table[name]; ok {
				js.AnyOf = append(js.AnyOf, named.Convert())
			} else {
				jsType := getJsonType(str)
				js.Type = append(js.Type, jsType)
				if str == "int" {
					js.Format = "int32"
				}
			}
		} else if wrapped, ok := s.Type[0].(*Schema); ok {
			js.AnyOf = append(js.AnyOf, wrapped.Convert())
		}
	}

	if len(s.Fields) > 0 {
		js.Properties = &json.Schemas{LinkedHashMap: sortedmap.LinkedHashMap[string, *json.Schema]{}}
		for _, f := range s.Fields {
			js.Properties.Set(f.Name, f.Convert())
		}
	}

	for _, symbol := range s.Symbols {
		js.Enum = append(js.Enum, symbol)
	}

	if s.Items != nil {
		js.Items = s.Items.Convert()
	}

	if s.Values != nil {
		js.AdditionalProperties = s.Values.Convert()
	}

	if len(s.Type) == 1 && s.Type[0] == "fixed" {
		js.MinLength = &s.Size
		js.MaxLength = &s.Size
	}

	return js
}

func getJsonType(t string) string {
	switch t {
	case "boolean":
		return t
	case "int", "long":
		return "integer"
	case "float", "double":
		return "number"
	case "record":
		return "object"
	case "enum":
		return "string"
	case "array":
		return "array"
	case "map":
		return "object"
	case "fixed":
		return "string"
	case "string":
		return "string"
	case "null":
		return "null"
	default:
		return t
	}
}
