package schema_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestJson_Structuring(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "JSON pointer",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": dynamictest.NewConfig(schematest.New("object",
							schematest.WithProperty("street_address", schematest.New("string")),
						)),
					},
				}
				r := &dynamic.Reference[schema.Schema]{Ref: "https://example.com/schemas/address#/properties/street_address"}
				s, err := r.Resolve(dynamictest.NewConfig(&schema.Schema{}), reader)
				require.NoError(t, err)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "$anchor",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": dynamictest.NewConfig(schematest.New("object",
							schematest.WithProperty("street_address",
								schematest.New("string", schematest.WithAnchor("street_address"))),
						), dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/address")))),
					},
				}

				person := dynamictest.NewConfig(&schema.Schema{Reference: dynamic.Reference[*schema.Schema]{Ref: "https://example.com/schemas/address#street_address"}}, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/person"))))
				person.OpenScope("")

				err := person.Data().(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data().(*schema.Schema).Type.String())
			},
		},
		{
			name: "$anchor with $id",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schema/billing-address": dynamictest.NewConfig(schematest.New("object",
							schematest.WithId("https://example.com/schema/address"),
							schematest.WithProperty("street_address",
								schematest.New("string", schematest.WithAnchor("street_address"))),
						), dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schema/billing-address")))),
					},
				}

				person := dynamictest.NewConfig(&schema.Schema{Reference: dynamic.Reference[*schema.Schema]{Ref: "https://example.com/schema/billing-address#street_address"}}, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schema/billing-address"))))

				err := person.Data().(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data().(*schema.Schema).Type.String())
			},
		},
		{
			name: "$anchor not in same scope",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{}

				person := dynamictest.NewConfig(schematest.New("object",
					schematest.WithId("https://example.com/schemas/person"),
					schematest.WithPropertyRef("foo", "#foo"),
					schematest.WithDef("foo",
						schematest.New("string",
							schematest.WithId("https://example.com/schemas/foo"),
							schematest.WithAnchor("foo"),
						),
					),
				), dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/person"))))

				err := person.Data().(*schema.Schema).Parse(person, reader)

				require.EqualError(t, err, "parse schema 'foo' failed: resolve reference '#foo' failed: name 'foo' not found in scope 'https://example.com/schemas/person'")
			},
		},
		{
			name: "relative to $id",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": dynamictest.NewConfig(schematest.New("object",
							schematest.WithProperty("street_address", schematest.New("string")),
						)),
					},
				}

				cfg := dynamictest.NewConfig(&schema.Schema{Id: "https://example.com/schemas/customer"})

				r := &dynamic.Reference[schema.Schema]{Ref: "/schemas/address"}
				s, err := r.Resolve(cfg, reader)
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Equal(t, "object", s.Type.String())
			},
		},
		{
			name: "$defs",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithPropertyRef("first_name", "#/$defs/name"),
					schematest.WithDef("name", schematest.New("string")),
				)

				err := s.Parse(dynamictest.NewConfig(s), &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "string", s.Properties.Get("first_name").Type.String())
			},
		},
		{
			name: "$ref using in referenced file",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"/bar.json": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("/bar.json")),
							Raw: []byte(`{
"$defs": { "items": { "type": "integer" } },
"type": "array",
"items": { "$ref": "#/$defs/items" }
}`),
						},
					},
				}

				foo := dynamictest.NewConfig(&schema.Schema{
					Reference: dynamic.Reference[*schema.Schema]{Ref: "/bar.json"},
				}, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("/foo.json"))))

				err := foo.Data().(*schema.Schema).Parse(foo, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "integer", foo.Data().(*schema.Schema).Items.Type.String())
			},
		},
		{
			name: "recursion",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("children",
						schematest.New("array", schematest.WithItemsRef("#")),
					),
				)

				err := s.Parse(dynamictest.NewConfig(s), &dynamictest.Reader{})

				require.NoError(t, err)
				children := s.Properties.Get("children")
				require.Equal(t, s, children.Items.Sub)
			},
		},
		{
			name: "recursion using $def",
			test: func(t *testing.T) {
				data := `
$defs:
  a:
    $ref: #/$defs/b
  b:
    $ref: '#'
$ref: '#/$defs/a'
`
				var s *schema.Schema
				err := yaml.Unmarshal([]byte(data), &s)
				require.NoError(t, err)

				err = s.Parse(dynamictest.NewConfig(s), &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "empty schema", s.String())
			},
		},
		{
			name: "self-recursion",
			test: func(t *testing.T) {
				data := `
$defs:
  a:
    properties:
      part:
        $ref: '#/$defs/a'
$ref: '#/$defs/a'
`
				var s *schema.Schema
				err := yaml.Unmarshal([]byte(data), &s)
				require.NoError(t, err)

				err = s.Parse(dynamictest.NewConfig(s), &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "schema properties=[part]", s.String())
			},
		},
		{
			name: "generic list of strings",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/list-of-t": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/list-of-t")),
							Raw: []byte(`{
"$defs": { "content": { "$dynamicAnchor": "T", "not": true } },
"type": "array",
"items": { "$dynamicRef": "#T" }
}`),
						},
					},
				}

				person := dynamictest.NewConfig(&schema.Schema{
					Defs: map[string]*schema.Schema{
						"string-items": {
							DynamicAnchor: "T",
							Type:          jsonSchema.Types{"string"},
						},
					},
					Reference: dynamic.Reference[*schema.Schema]{Ref: "https://example.com/schemas/list-of-t"},
				}, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/list-of-string"))))

				err := person.Data().(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data().(*schema.Schema).Items.Type.String())
			},
		},
		{
			name: "parsing twice with $refs",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/foo": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/foo")),
							Raw: []byte(`{
"$defs": { "items": { "type": "integer" } },
"type": "array",
"items": { "$ref": "#/$defs/items" },
}`),
						},
					},
				}

				person := dynamictest.NewConfig(&schema.Schema{
					Reference: dynamic.Reference[*schema.Schema]{Ref: "https://example.com/schemas/foo"},
				}, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/bar"))))

				// 1. parsing
				err := person.Data().(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				// 2. parsing
				err = person.Data().(*schema.Schema).Parse(person, reader)

				require.NoError(t, err)
				require.Equal(t, "integer", person.Data().(*schema.Schema).Items.Type.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
