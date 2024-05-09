package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mokapi/json/generator"
	jsonSchema "mokapi/json/schema"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/sortedmap"
	"net/http"
)

type Properties struct {
	sortedmap.LinkedHashMap[string, *schemaInfo]
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
	Description string `json:"description,omitempty"`
	Ref         string `json:"ref,omitempty"`

	Type       interface{}   `json:"type"`
	AnyOf      []*schemaInfo `json:"anyOf,omitempty"`
	AllOf      []*schemaInfo `json:"allOf,omitempty"`
	OneOf      []*schemaInfo `json:"oneOf,omitempty"`
	Deprecated bool          `json:"deprecated,omitempty"`
	Example    interface{}   `json:"example,omitempty"`
	Enum       []interface{} `json:"enum,omitempty"`
	Xml        *xml          `json:"xml,omitempty"`
	Format     string        `json:"format,omitempty"`
	Nullable   bool          `json:"nullable,omitempty"`

	Pattern   string `json:"pattern,omitempty"`
	MinLength *int   `yaml:"minLength" json:"minLength,omitempty"`
	MaxLength *int   `yaml:"maxLength" json:"maxLength,omitempty"`

	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *bool    `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *bool    `json:"exclusiveMaximum,omitempty"`

	Items        *schemaInfo `json:"items,omitempty"`
	UniqueItems  bool        `json:"uniqueItems,omitempty"`
	MinItems     *int        `json:"minItems,omitempty"`
	MaxItems     *int        `json:"maxItems,omitempty"`
	ShuffleItems bool        `json:"shuffleItems,omitempty"`

	Properties           *Properties `json:"properties,omitempty"`
	Required             []string    `json:"required,omitempty"`
	AdditionalProperties interface{} `json:"additionalProperties,omitempty"`
	MinProperties        *int        `json:"minProperties,omitempty"`
	MaxProperties        *int        `json:"maxProperties,omitempty"`
}

type xml struct {
	Wrapped   bool   `json:"wrapped,omitempty"`
	Name      string `json:"name,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type additionalProperties struct {
	Schema    *schemaInfo `json:"schema,omitempty"`
	Forbidden bool        `json:"forbidden,omitempty"`
}

type requestExample struct {
	Name   string      `json:"name,omitempty"`
	Schema *schemaInfo `json:"schema,omitempty"`
}

func (h *handler) getExampleData(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if len(accept) == 0 {
		accept = "application/json"
	}
	ct := media.ParseContentType(accept)
	if ct.IsAny() {
		ct = media.ParseContentType("application/json")
	}
	if ct.Subtype != "json" && ct.Subtype != "xml" {
		http.Error(w, fmt.Sprintf("Content type %v not supported. Only json or xml are supported", ct), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", ct.String())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	re := &requestExample{}
	err = json.Unmarshal(body, &re)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s := toSchema(re.Schema)
	data, err := generator.New(&generator.Request{
		Path: generator.Path{
			&generator.PathElement{Name: re.Name, Schema: schema.ConvertToJsonSchema(&schema.Ref{Value: s})},
		},
	})
	//data, err := schema.CreateValue(&schema.Ref{Value: s})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b, err := s.Marshal(data, ct)
	if err != nil {
		writeError(w, fmt.Errorf("failed request %v: %w", r.URL.String(), err), http.StatusInternalServerError)
	}
	_, err = w.Write(b)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}
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
	defer func() {
		delete(c.schemas, s.Ref)
	}()

	result := &schemaInfo{
		Description: s.Value.Description,
		Ref:         s.Ref,

		Type:    s.Value.Type,
		Example: s.Value.Example,
		Enum:    s.Value.Enum,
		Format:  s.Value.Format,

		Pattern:   s.Value.Pattern,
		MinLength: s.Value.MinLength,
		MaxLength: s.Value.MaxLength,

		Minimum:          s.Value.Minimum,
		Maximum:          s.Value.Maximum,
		ExclusiveMinimum: s.Value.ExclusiveMinimum,
		ExclusiveMaximum: s.Value.ExclusiveMaximum,

		UniqueItems:  s.Value.UniqueItems,
		MinItems:     s.Value.MinItems,
		MaxItems:     s.Value.MaxItems,
		ShuffleItems: s.Value.ShuffleItems,

		Required:      s.Value.Required,
		MinProperties: s.Value.MinProperties,
		MaxProperties: s.Value.MaxProperties,
	}

	if len(s.Value.Type) == 0 {
		result.Type = ""
	} else if len(s.Value.Type) == 1 {
		result.Type = s.Value.Type[0]
	}

	if s.Value.Nullable && !s.Value.Type.IsNullable() {
		result.Type = append(s.Value.Type, "null")
	}

	if len(s.Ref) > 0 {
		c.schemas[s.Ref] = result
	}

	result.Items = c.getSchema(s.Value.Items)

	if s.Value.Properties != nil {
		result.Properties = &Properties{}
		for it := s.Value.Properties.Iter(); it.Next(); {
			key := it.Key()
			_ = key
			prop := c.getSchema(it.Value())
			if prop == nil {
				continue
			}
			result.Properties.Set(it.Key(), prop)
		}
	}

	if s.Value.Xml != nil {
		result.Xml = &xml{
			Wrapped:   s.Value.Xml.Wrapped,
			Name:      s.Value.Xml.Name,
			Attribute: s.Value.Xml.Attribute,
			Prefix:    s.Value.Xml.Prefix,
			Namespace: s.Value.Xml.Namespace,
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

	if s.Value.AdditionalProperties != nil {
		if s.Value.AdditionalProperties.Forbidden {
			result.AdditionalProperties = !s.Value.AdditionalProperties.Forbidden
		} else if s.Value.AdditionalProperties.Ref != nil {
			result.AdditionalProperties = getSchema(s.Value.AdditionalProperties.Ref)
		}
	}

	return result
}

func toSchema(s *schemaInfo) *schema.Schema {
	if s == nil {
		return nil
	}
	result := &schema.Schema{
		Description: s.Description,

		Deprecated: s.Deprecated,
		Example:    s.Example,
		Enum:       s.Enum,
		Format:     s.Format,
		Nullable:   s.Nullable,

		Pattern:   s.Pattern,
		MinLength: s.MinLength,
		MaxLength: s.MaxLength,

		Minimum:          s.Minimum,
		Maximum:          s.Maximum,
		ExclusiveMinimum: s.ExclusiveMinimum,
		ExclusiveMaximum: s.ExclusiveMaximum,

		UniqueItems:  s.UniqueItems,
		MinItems:     s.MinItems,
		MaxItems:     s.MaxItems,
		ShuffleItems: s.ShuffleItems,

		Required:      s.Required,
		MinProperties: s.MinProperties,
		MaxProperties: s.MaxProperties,
	}

	if s.Type != nil {
		switch v := s.Type.(type) {
		case string:
			result.Type = jsonSchema.Types{fmt.Sprintf("%v", s.Type)}
		case []interface{}:
			for _, typeName := range v {
				result.Type = append(result.Type, fmt.Sprintf("%v", typeName))
			}
		}
	}

	if s.Properties != nil && s.Properties.Len() > 0 {
		result.Properties = &schema.Schemas{}
		for it := s.Properties.Iter(); it.Next(); {
			result.Properties.Set(it.Key(), &schema.Ref{Value: toSchema(it.Value())})
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

	if s.AdditionalProperties != nil {
		if b, ok := s.AdditionalProperties.(bool); ok {
			result.AdditionalProperties = &schema.AdditionalProperties{Forbidden: !b}
		} else if additional, ok := s.AdditionalProperties.(*schemaInfo); ok {
			result.AdditionalProperties.Ref = &schema.Ref{Value: toSchema(additional)}
		}
	}

	return result
}
