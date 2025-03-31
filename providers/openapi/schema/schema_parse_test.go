package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
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
						"https://example.com/schemas/address": {
							Data: schematest.New("object",
								schematest.WithProperty("street_address", schematest.New("string")),
							),
						},
					},
				}
				r := &schema.Schema{}
				err := dynamic.Resolve("https://example.com/schemas/address#/properties/street_address", &r, &dynamic.Config{Data: &schema.Schema{}}, reader)
				require.NoError(t, err)
				require.Equal(t, "string", r.Type.String())
			},
		},
		{
			name: "$anchor",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/address")),
							Data: schematest.New("object",
								schematest.WithProperty("street_address",
									schematest.New("string", schematest.WithAnchor("street_address"))),
							),
						},
					},
				}

				person := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/person")),
					Data: &schema.Schema{Ref: "https://example.com/schemas/address#street_address"},
				}
				person.OpenScope("")

				err := person.Data.(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data.(*schema.Schema).Type.String())
			},
		},
		{
			name: "$anchor with $id",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schema/billing-address": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schema/billing-address")),
							Data: schematest.New("object",
								schematest.WithId("https://example.com/schema/address"),
								schematest.WithProperty("street_address",
									schematest.New("string", schematest.WithAnchor("street_address"))),
							),
						},
					},
				}

				person := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schema/billing-address")),
					Data: &schema.Schema{Ref: "https://example.com/schema/billing-address#street_address"},
				}

				err := person.Data.(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data.(*schema.Schema).Type.String())
			},
		},
		{
			name: "$anchor not in same scope",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{}

				person := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/person")),
					Data: schematest.New("object",
						schematest.WithId("https://example.com/schemas/person"),
						schematest.WithPropertyRef("foo", "#foo"),
						schematest.WithDef("foo",
							schematest.New("string",
								schematest.WithId("https://example.com/schemas/foo"),
								schematest.WithAnchor("foo"),
							),
						),
					),
				}

				err := person.Data.(*schema.Schema).Parse(person, reader)

				require.EqualError(t, err, "parse schema 'foo' failed: resolve reference '#foo' failed: name 'foo' not found in scope 'https://example.com/schemas/person'")
			},
		},
		{
			name: "relative to $id",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": {
							Data: schematest.New("object",
								schematest.WithProperty("street_address", schematest.New("string")),
							),
						},
					},
				}

				cfg := &dynamic.Config{Data: &schema.Schema{Id: "https://example.com/schemas/customer"}}

				r := &schema.Schema{}
				err := dynamic.Resolve("/schemas/address", &r, cfg, reader)
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Equal(t, "object", r.Type.String())
			},
		},
		{
			name: "$defs",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithPropertyRef("first_name", "#/$defs/name"),
					schematest.WithDef("name", schematest.New("string")),
				)

				err := s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "string", s.Properties.Get("first_name").Type.String())
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

				err := s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				children := s.Properties.Get("children")
				require.Equal(t, s, children.Items.Sub)
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

				person := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/list-of-string")),
					Data: &schema.Schema{
						Defs: map[string]*schema.Schema{
							"string-items": {
								DynamicAnchor: "T",
								Type:          jsonSchema.Types{"string"},
							},
						},
						Ref: "https://example.com/schemas/list-of-t",
					},
				}

				err := person.Data.(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data.(*schema.Schema).Items.Type.String())
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
