package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/sortedmap"
	"net/http"
)

type Properties struct {
	sortedmap.LinkedHashMap
}

func (p *Properties) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("unexpected token %s; expected '{'", token)
	}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		s := &schemaInfo{}
		err = dec.Decode(&s)
		if err != nil {
			return err
		}
		p.Set(key, s)
	}
}

type schemaInfo struct {
	Name             string        `json:"name,omitempty"`
	Description      string        `json:"description,omitempty"`
	Ref              string        `json:"ref,omitempty"`
	Type             string        `json:"type"`
	Properties       *Properties   `json:"properties,omitempty"`
	Required         []string      `json:"required,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
	Items            *schemaInfo   `json:"items,omitempty"`
	Format           string        `json:"format,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	Xml              *xml          `json:"xml,omitempty"`
	Nullable         bool          `json:"nullable,omitempty"`
	Example          interface{}   `json:"example,omitempty"`
	Minimum          *float64      `json:"minimum,omitempty"`
	Maximum          *float64      `json:"maximum,omitempty"`
	ExclusiveMinimum *bool         `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *bool         `json:"exclusiveMaximum,omitempty"`
	AnyOf            []*schemaInfo `json:"anyOf,omitempty"`
	AllOf            []*schemaInfo `json:"allOf,omitempty"`
	OneOf            []*schemaInfo `json:"oneOf,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty"`
	MinItems         *int          `json:"minItems,omitempty"`
	MaxItems         *int          `json:"maxItems,omitempty"`
	ShuffleItems     bool          `json:"shuffleItems,omitempty"`
	MinProperties    *int          `json:"minProperties,omitempty"`
	MaxProperties    *int          `json:"maxProperties,omitempty"`
}

type xml struct {
	Wrapped   bool   `json:"wrapped,omitempty"`
	Name      string `json:"name,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	CData     bool   `json:"x-cdata"`
}

func (h *handler) getExampleData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	si := &schemaInfo{}
	err = json.Unmarshal(body, &si)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s := toSchema(si)
	data := schema.NewGenerator().New(&schema.Ref{Value: s})

	writeJsonBody(w, data)
}

func getSchema(s *schema.Ref) *schemaInfo {
	converter := &schemaConverter{map[string]*schemaInfo{}}
	return converter.getSchema(s)
}

type schemaConverter struct {
	schemas map[string]*schemaInfo
}

func (c *schemaConverter) getSchema(s *schema.Ref) *schemaInfo {
	if s == nil || s.Value == nil {
		return nil
	}

	// loop protection, only return reference
	if _, ok := c.schemas[s.Ref]; ok {
		return &schemaInfo{Ref: s.Ref}
	}

	result := &schemaInfo{
		Description:      s.Value.Description,
		Ref:              s.Ref,
		Type:             s.Value.Type,
		Enum:             s.Value.Enum,
		Required:         s.Value.Required,
		Format:           s.Value.Format,
		Pattern:          s.Value.Pattern,
		Nullable:         s.Value.Nullable,
		Example:          s.Value.Example,
		Minimum:          s.Value.Minimum,
		Maximum:          s.Value.Maximum,
		ExclusiveMinimum: s.Value.ExclusiveMinimum,
		ExclusiveMaximum: s.Value.ExclusiveMaximum,
		UniqueItems:      s.Value.UniqueItems,
		MinItems:         s.Value.MinItems,
		MaxItems:         s.Value.MaxItems,
		ShuffleItems:     s.Value.ShuffleItems,
		MinProperties:    s.Value.MinProperties,
		MaxProperties:    s.Value.MaxProperties,
	}

	if len(s.Ref) > 0 {
		c.schemas[s.Ref] = result
	}

	result.Items = c.getSchema(s.Value.Items)

	if s.Value.Properties != nil && s.Value.Properties.Value != nil {
		result.Properties = &Properties{}
		for it := s.Value.Properties.Value.Iter(); it.Next(); {
			prop := c.getSchema(it.Value().(*schema.Ref))
			prop.Name = it.Key().(string)
			result.Properties.Set(prop.Name, prop)
		}
	}

	if s.Value.Xml != nil {
		result.Xml = &xml{
			Wrapped:   s.Value.Xml.Wrapped,
			Name:      s.Value.Xml.Name,
			Attribute: s.Value.Xml.Attribute,
			Prefix:    s.Value.Xml.Prefix,
			Namespace: s.Value.Xml.Namespace,
			CData:     s.Value.Xml.CData,
		}
	}

	for _, any := range s.Value.AnyOf {
		result.AnyOf = append(result.AnyOf, c.getSchema(any))
	}
	for _, all := range s.Value.AllOf {
		result.AllOf = append(result.AnyOf, c.getSchema(all))
	}
	for _, one := range s.Value.OneOf {
		result.OneOf = append(result.AnyOf, c.getSchema(one))
	}

	return result
}

func toSchema(s *schemaInfo) *schema.Schema {
	if s == nil {
		return nil
	}
	result := &schema.Schema{
		Type:             s.Type,
		Format:           s.Format,
		Pattern:          s.Pattern,
		Description:      s.Description,
		Nullable:         s.Nullable,
		Example:          s.Example,
		Required:         s.Required,
		Enum:             s.Enum,
		Minimum:          s.Minimum,
		Maximum:          s.Maximum,
		ExclusiveMinimum: s.ExclusiveMinimum,
		ExclusiveMaximum: s.ExclusiveMaximum,
		UniqueItems:      s.UniqueItems,
		MinItems:         s.MinItems,
		MaxItems:         s.MaxItems,
		ShuffleItems:     s.ShuffleItems,
		MinProperties:    s.MinProperties,
		MaxProperties:    s.MaxProperties,
	}
	if s.Properties != nil && s.Properties.Len() > 0 {
		result.Properties = &schema.SchemasRef{Value: &schema.Schemas{}}
		for it := s.Properties.Iter(); it.Next(); {
			result.Properties.Value.Set(it.Key(), &schema.Ref{Value: toSchema(it.Value().(*schemaInfo))})
		}
	}
	if s.Items != nil {
		result.Items = &schema.Ref{Value: toSchema(s.Items)}
	}
	if s.Xml != nil {
		result.Xml = &schema.Xml{
			Wrapped:   s.Xml.Wrapped,
			Name:      s.Xml.Name,
			Attribute: s.Xml.Attribute,
			Prefix:    s.Xml.Prefix,
			Namespace: s.Xml.Namespace,
			CData:     s.Xml.CData,
		}
	}
	for _, any := range s.AnyOf {
		result.AnyOf = append(result.AnyOf, &schema.Ref{Value: toSchema(any)})
	}
	for _, all := range s.AllOf {
		result.AllOf = append(result.AllOf, &schema.Ref{Value: toSchema(all)})
	}
	for _, one := range s.OneOf {
		result.OneOf = append(result.OneOf, &schema.Ref{Value: toSchema(one)})
	}
	return result
}
