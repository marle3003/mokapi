package openapi_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/version"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// maximalConfig returns a config in which every exported field that Config.Patch
// has to handle is set (see requireAllFieldsSet).
//
// With bare=true it returns the same structure, but without the values:
// maps and pointers that can be merged exist, identifiers (map keys, server URLs,
// parameter names, tag names, security scheme types) are kept, and everything else
// is empty. Patching the full config into the bare one has to give the full config.
// Slices and map entries that are missing in the bare config are adopted by Patch.
func maximalConfig(bare bool) *openapi.Config {
	c := &openapi.Config{
		OpenApi: ifFull(bare, version.New("3.1.0")),
		Info: openapi.Info{
			Name:           text(bare, "title"),
			Summary:        text(bare, "summary"),
			Description:    text(bare, "description"),
			Contact:        &openapi.Contact{Name: text(bare, "contact"), Url: text(bare, "https://contact.io"), Email: text(bare, "contact@foo.bar")},
			Version:        text(bare, "1.0"),
			TermsOfService: text(bare, "https://terms.io"),
			License:        &openapi.License{Name: text(bare, "MIT"), Url: text(bare, "https://license.io")},
		},
		Servers:      []*openapi.Server{{Url: "http://localhost:8080", Description: text(bare, "server")}},
		Paths:        pathItems(bare),
		Webhooks:     map[string]*openapi.PathRef{"hook": {Value: path(bare)}},
		Security:     ifFull(bare, []openapi.SecurityRequirement{{"oauth": {"read"}}}),
		Components:   components(bare),
		ExternalDocs: externalDocs(bare),
		Tags: []*openapi.Tag{{
			Name:         "tag",
			Summary:      text(bare, "tag summary"),
			Description:  text(bare, "tag description"),
			ExternalDocs: ifFull(bare, externalDocs(false)),
			Parent:       text(bare, "parent"),
			Kind:         text(bare, "nav"),
		}},
	}
	return c
}

func externalDocs(bare bool) *openapi.ExternalDocs {
	return &openapi.ExternalDocs{Description: text(bare, "docs"), Url: text(bare, "https://docs.io")}
}

func pathItems(bare bool) *openapi.PathItems {
	items := &openapi.PathItems{}
	items.Set("/foo", &openapi.PathRef{Value: pathWithName(bare, "/foo")})
	if !bare {
		items.Set("/ref", &openapi.PathRef{Reference: reference[*openapi.PathRef](), Value: &openapi.Path{Path: "/ref", Summary: "referenced"}})
	}
	return items
}

// pathWithName sets the name Parse derives from the key in Paths. Webhooks have none.
func pathWithName(bare bool, name string) *openapi.Path {
	p := path(bare)
	p.Path = name
	return p
}

// path keeps the derived fields (status, errors, operation method and path), also with bare=true:
// Patch does not merge them into an existing path, so they have to be equal.
func path(bare bool) *openapi.Path {
	p := &openapi.Path{
		Status:      openapi.StatusValidationError,
		Errors:      []openapi.Error{{Message: "path error"}},
		Summary:     text(bare, "summary"),
		Description: text(bare, "description"),
		Get:         operation(bare),
		Post:        operation(bare),
		Put:         operation(bare),
		Patch:       operation(bare),
		Delete:      operation(bare),
		Head:        operation(bare),
		Options:     operation(bare),
		Trace:       operation(bare),
		Query:       operation(bare),
		AdditionalOperations: map[string]*openapi.Operation{
			"LINK": operation(bare),
		},
		Parameters:  parameters(bare),
		MethodOrder: ifFull(bare, []string{"GET", "POST", "LINK"}),
	}
	for method, op := range p.Operations() {
		op.Method = method
		op.Path = p
	}
	return p
}

func operation(bare bool) *openapi.Operation {
	responses := &openapi.Responses{}
	responses.Set("200", &openapi.ResponseRef{Value: response(bare)})
	return &openapi.Operation{
		Status:      openapi.StatusInvalid,
		Errors:      []openapi.Error{{Message: "operation error"}},
		Tags:        ifFull(bare, []string{"pets"}),
		Summary:     text(bare, "summary"),
		Description: text(bare, "description"),
		Deprecated:  !bare,
		OperationId: text(bare, "operationId"),
		Parameters:  parameters(bare),
		RequestBody: &openapi.RequestBodyRef{Value: requestBody(bare)},
		Responses:   responses,
		Security:    ifFull(bare, []openapi.SecurityRequirement{{"basic": {}}}),
	}
}

// parameters keeps the name, because parameters are merged by name
func parameters(bare bool) openapi.Parameters {
	p := parameter(bare, openapi.ParameterQuery)
	p.Name = "id"
	return openapi.Parameters{{Value: p}}
}

func parameter(bare bool, location openapi.Location) *openapi.Parameter {
	p := parameterWithoutContent(bare, location)
	p.Content = openapi.Content{"application/json": mediaType(bare, "application/json", false)}
	return p
}

// parameterWithoutContent ends the recursion parameter -> media type -> encoding header -> parameter
func parameterWithoutContent(bare bool, location openapi.Location) *openapi.Parameter {
	return &openapi.Parameter{
		Name:          text(bare, "id"),
		Type:          ifFull(bare, location),
		Schema:        schemaWithAllKeywords(bare),
		Required:      pointer(bare, true),
		Description:   text(bare, "description"),
		Deprecated:    pointer(bare, true),
		Style:         text(bare, "form"),
		Explode:       pointer(bare, true),
		AllowReserved: pointer(bare, true),
	}
}

func header(bare bool) *openapi.Header {
	return &openapi.Header{Parameter: *parameter(bare, openapi.ParameterHeader)}
}

func requestBody(bare bool) *openapi.RequestBody {
	return &openapi.RequestBody{
		Description: text(bare, "description"),
		Content: openapi.Content{
			"application/json": mediaType(bare, "application/json", false),
			"application/xml":  mediaType(bare, "application/xml", true),
		},
		Required: pointer(bare, true),
	}
}

func response(bare bool) *openapi.Response {
	return &openapi.Response{
		Description: text(bare, "description"),
		Content: openapi.Content{
			"application/json": mediaType(bare, "application/json", false),
		},
		Headers: openapi.Headers{"X-Rate": {Value: header(bare)}},
	}
}

// mediaType sets either Example or Examples, because patching treats them as exclusive.
func mediaType(bare bool, contentType string, withExamples bool) *openapi.MediaType {
	m := &openapi.MediaType{
		Schema:      schemaWithAllKeywords(bare),
		ContentType: ifFull(bare, media.ParseContentType(contentType)),
		Encoding: map[string]*openapi.Encoding{
			"file": {
				ContentType:   text(bare, "image/png"),
				Headers:       map[string]openapi.Header{"X-Part": {Parameter: *parameterWithoutContent(bare, openapi.ParameterHeader)}},
				Style:         text(bare, "form"),
				Explode:       !bare,
				AllowReserved: !bare,
			},
		},
	}
	if bare {
		return m
	}
	if withExamples {
		m.Examples = openapi.Examples{"e": {Value: example()}}
	} else {
		m.Example = &openapi.ExampleValue{Value: "foo"}
	}
	return m
}

func example() *openapi.Example {
	return &openapi.Example{Summary: "summary", Description: "description", Value: "foo", ExternalValue: "https://example.io/foo"}
}

func components(bare bool) openapi.Components {
	schemas := &schema.Schemas{}
	schemas.Set("Pet", schemaWithAllKeywords(bare))

	c := openapi.Components{
		Schemas:       schemas,
		Responses:     openapi.ResponseBodies{"r": {Value: response(bare)}},
		RequestBodies: openapi.RequestBodies{"rb": {Value: requestBody(bare)}},
		Parameters:    openapi.ComponentParameters{"p": {Value: parameter(bare, openapi.ParameterQuery)}},
		Examples:      openapi.Examples{},
		Headers:       openapi.Headers{"h": {Value: header(bare)}},
		PathItems:     pathItems(bare),
		SecuritySchemes: openapi.SecuritySchemes{
			"basic": &openapi.HttpSecurityScheme{
				Type:         "http",
				Scheme:       text(bare, "bearer"),
				BearerFormat: text(bare, "JWT"),
				Description:  text(bare, "description"),
			},
			"key": &openapi.ApiKeySecurityScheme{
				Type:        "apiKey",
				In:          text(bare, "header"),
				Name:        text(bare, "X-Key"),
				Description: text(bare, "description"),
			},
			"oauth": &openapi.OAuth2SecurityScheme{
				Type:        "oauth2",
				Description: text(bare, "description"),
				Flows: map[string]*openapi.OAuth2Flow{"implicit": {
					AuthorizationUrl: text(bare, "https://auth.io"),
					TokenUrl:         text(bare, "https://token.io"),
					RefreshUrl:       text(bare, "https://refresh.io"),
					Scopes:           map[string]string{"read": text(bare, "read access")},
				}},
			},
			"other": &openapi.NotSupportedSecurityScheme{Type: "mutualTLS"},
		},
	}
	if !bare {
		// references are adopted as a whole, they only exist next to a merged entry
		c.Schemas.Set("Ref", &schema.Schema{Reference: reference[*schema.Schema](), Type: jsonSchema.Types{"object"}})
		c.Responses["ref"] = &openapi.ResponseRef{Reference: reference[*openapi.ResponseRef](), Value: response(false)}
		c.RequestBodies["ref"] = &openapi.RequestBodyRef{Reference: reference[*openapi.RequestBodyRef](), Value: requestBody(false)}
		c.Parameters["ref"] = &openapi.ParameterRef{Reference: reference[*openapi.ParameterRef](), Value: parameter(false, openapi.ParameterQuery)}
		c.Examples["ref"] = &openapi.ExampleRef{Reference: reference[*openapi.ExampleRef](), Value: example()}
		c.Headers["ref"] = &openapi.HeaderRef{Reference: reference[*openapi.HeaderRef](), Value: header(false)}
	}
	return c
}

// schemaWithAllKeywords sets every keyword of schema.Schema. Nested schemas are
// simple, so the fixture has no cycles.
func schemaWithAllKeywords(bare bool) *schema.Schema {
	nested := func() *schema.Schema {
		if bare {
			return &schema.Schema{}
		}
		return &schema.Schema{Type: jsonSchema.Types{"string"}}
	}
	nestedList := func() []*schema.Schema {
		if bare {
			return nil
		}
		return []*schema.Schema{{Title: "nested", Type: jsonSchema.Types{"string"}}}
	}
	s := &schema.Schema{
		Id:            text(bare, "https://schema.io/pet"),
		Schema:        text(bare, "https://json-schema.org/draft/2020-12/schema"),
		Boolean:       pointer(bare, true),
		Anchor:        text(bare, "anchor"),
		DynamicAnchor: text(bare, "dynamic"),

		Type:  ifFull(bare, jsonSchema.Types{"object", "null"}),
		Enum:  ifFull(bare, []any{"a", "b"}),
		Const: pointer[any](bare, "a"),

		MultipleOf:       pointer(bare, 2.0),
		Minimum:          pointer(bare, 1.0),
		Maximum:          pointer(bare, 10.0),
		ExclusiveMinimum: ifFull(bare, jsonSchema.NewUnionTypeA[float64, bool](1.5)),
		ExclusiveMaximum: ifFull(bare, jsonSchema.NewUnionTypeB[float64, bool](true)),

		Pattern:   text(bare, "^[a-z]+$"),
		MinLength: pointer(bare, 1),
		MaxLength: pointer(bare, 10),
		Format:    text(bare, "email"),

		Items:            nested(),
		PrefixItems:      nestedList(),
		UnevaluatedItems: nested(),
		Contains:         nested(),
		MaxContains:      pointer(bare, 5),
		MinContains:      pointer(bare, 1),
		MinItems:         pointer(bare, 1),
		MaxItems:         pointer(bare, 5),
		UniqueItems:      pointer(bare, true),
		ShuffleItems:     !bare,

		PatternProperties:     map[string]*schema.Schema{"^x-": nested()},
		MinProperties:         pointer(bare, 1),
		MaxProperties:         pointer(bare, 5),
		Required:              ifFull(bare, []string{"name"}),
		DependentRequired:     map[string][]string{"name": ifFull(bare, []string{"id"})},
		DependentSchemas:      map[string]*schema.Schema{"name": nested()},
		AdditionalProperties:  nested(),
		UnevaluatedProperties: nested(),
		PropertyNames:         nested(),

		AnyOf: nestedList(),
		AllOf: nestedList(),
		OneOf: nestedList(),
		Not:   nested(),
		If:    nested(),
		Then:  nested(),
		Else:  nested(),

		Title:       text(bare, "Pet"),
		Description: text(bare, "A pet"),
		Default:     ifFull[any](bare, "a"),
		Deprecated:  !bare,
		Examples:    ifFull(bare, []jsonSchema.Example{{Value: "a"}}),
		Example:     ifFull(bare, &jsonSchema.Example{Value: "b"}),

		ContentMediaType: text(bare, "application/json"),
		ContentEncoding:  text(bare, "base64"),

		Xml:      &schema.Xml{Name: text(bare, "pet"), Prefix: text(bare, "p"), Namespace: text(bare, "https://xml.io"), Wrapped: !bare, Attribute: !bare},
		Nullable: !bare,

		Definitions: map[string]*schema.Schema{"d": nested()},
		Defs:        map[string]*schema.Schema{"d": nested()},
	}
	s.Properties = &schema.Schemas{}
	s.Properties.Set("name", nested())
	return s
}

func reference[T any]() dynamic.Reference[T] {
	return dynamic.Reference[T]{Ref: "#/components/foo", DynamicRef: "#foo", Summary: "ref summary", Description: "ref description"}
}

func text(bare bool, s string) string {
	if bare {
		return ""
	}
	return s
}

func pointer[T any](bare bool, v T) *T {
	if bare {
		return nil
	}
	return &v
}

func ifFull[T any](bare bool, v T) T {
	if bare {
		var zero T
		return zero
	}
	return v
}

// derivedFields are set by Parse or the runtime, not by the document, and are not patched.
// Operation.Path and Operation.Method are skipped because Path and Operation point at each other;
// TestConfig_Patch_OperationsPointAtTheirPath covers them.
var derivedFields = map[string]bool{
	"Operation.Path":   true,
	"Operation.Method": true,
	"Server.IsDefault": true,
	"Schema.Sub":       true,
}

// requireAllFieldsSet fails if an exported field of a struct type reachable from v
// is zero in every instance found. A field added later has to be added to the
// fixture, and therefore to Patch, too.
func requireAllFieldsSet(t *testing.T, v any) {
	t.Helper()
	c := &fieldCoverage{set: map[string]bool{}, all: map[string]bool{}}
	c.walk(reflect.ValueOf(v))

	var missing []string
	for name := range c.all {
		if !c.set[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	require.Empty(t, missing, "fields not set in the fixture")
}

type fieldCoverage struct {
	set map[string]bool
	all map[string]bool
}

func (c *fieldCoverage) walk(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			c.walk(v.Elem())
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			c.walk(v.Index(i))
		}
	case reflect.Map:
		for it := v.MapRange(); it.Next(); {
			c.walk(it.Value())
		}
	case reflect.Struct:
		c.walkStruct(v)
	}
}

func (c *fieldCoverage) walkStruct(v reflect.Value) {
	t := v.Type()
	if !strings.HasPrefix(t.PkgPath(), "mokapi/providers/openapi") && !strings.HasPrefix(t.Name(), "Reference[") {
		return
	}
	if !v.CanAddr() {
		// map values: methods with pointer receivers need an addressable copy
		cp := reflect.New(t).Elem()
		cp.Set(v)
		v = cp
	}
	if values := v.Addr().MethodByName("Values"); values.IsValid() {
		// sortedmap.LinkedHashMap keeps its entries private
		c.walk(values.Call(nil)[0])
		return
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := fmt.Sprintf("%s.%s", t.Name(), f.Name)
		if !f.IsExported() || derivedFields[name] {
			continue
		}
		c.all[name] = true
		if fv := v.Field(i); !fv.IsZero() {
			c.set[name] = true
			c.walk(fv)
		}
	}
}
