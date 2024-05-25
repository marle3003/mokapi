package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/schema/json/ref"
	"mokapi/sortedmap"
	"strings"
	"testing"
)

func TestMarshal_Xml(t *testing.T) {
	testcases := []struct {
		name   string
		data   func() interface{}
		schema *schema.Ref
		test   func(t *testing.T, s string, err error)
	}{
		{
			name: "no schema",
			data: func() interface{} {
				return 4
			},
			schema: nil,
			test: func(t *testing.T, s string, err error) {
				require.EqualError(t, err, "encoding data to 'application/xml' failed: no schema provided")
			},
		},
		{
			name: "no xml name",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer"),
			test: func(t *testing.T, s string, err error) {
				require.EqualError(t, err, "encoding data to 'application/xml' failed: root element name is undefined: reference name of schema and attribute xml.name is empty")
			},
		},
		{
			name: "root name from xml.name",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer", schematest.WithXml(&schema.Xml{Name: "foo"})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "<foo>4</foo>", s)
			},
		},
		{
			name: "root name from reference name",
			data: func() interface{} {
				return 4
			},
			schema: &schema.Ref{Reference: ref.Reference{Ref: "#/components/schemas/foo"}, Value: schematest.New("integer")},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "<foo>4</foo>", s)
			},
		},
		{
			name: "namespace",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer",
				schematest.WithXml(&schema.Xml{
					Name:      "foo",
					Namespace: "https://foo.bar",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo xmlns="https://foo.bar">4</foo>`, s)
			},
		},
		{
			name: "prefix",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer",
				schematest.WithXml(&schema.Xml{
					Name:      "foo",
					Prefix:    "ns",
					Namespace: "https://foo.bar",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<ns:foo xmlns:ns="https://foo.bar">4</ns:foo>`, s)
			},
		},
		{
			name: "prefix & namespace",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer",
				schematest.WithXml(&schema.Xml{
					Name:   "foo",
					Prefix: "ns",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<ns:foo>4</ns:foo>`, s)
			},
		},
		{
			name: "wrapped but not array",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer",
				schematest.WithXml(&schema.Xml{
					Name:    "foo",
					Wrapped: true,
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo>4</foo>`, s)
			},
		},
		{
			name: "string",
			data: func() interface{} {
				return "bar"
			},
			schema: schematest.NewRef("string",
				schematest.WithXml(&schema.Xml{
					Name: "foo",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo>bar</foo>`, s)
			},
		},
		{
			name: "string escape",
			data: func() interface{} {
				return "<bar>"
			},
			schema: schematest.NewRef("string",
				schematest.WithXml(&schema.Xml{
					Name: "foo",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo>&lt;bar&gt;</foo>`, s)
			},
		},
		{
			name: "number",
			data: func() interface{} {
				return 4.7
			},
			schema: schematest.NewRef("number",
				schematest.WithXml(&schema.Xml{
					Name: "foo",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo>4.7</foo>`, s)
			},
		},
		{
			name: "integer",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("integer",
				schematest.WithXml(&schema.Xml{
					Name: "foo",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo>4</foo>`, s)
			},
		},
		{
			name: "boolean",
			data: func() interface{} {
				return true
			},
			schema: schematest.NewRef("boolean",
				schematest.WithXml(&schema.Xml{
					Name: "foo",
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo>true</foo>`, s)
			},
		},
		{
			name: "array",
			data: func() interface{} {
				return []interface{}{1, 2, 3}
			},
			schema: schematest.NewRef("array",
				schematest.WithItems("integer"),
				schematest.WithXml(&schema.Xml{
					Name:    "foo",
					Wrapped: true, // without wrapped results in invalid xml when on root
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo><foo>1</foo><foo>2</foo><foo>3</foo></foo>`, s)
			},
		},
		{
			name: "array with different item name",
			data: func() interface{} {
				return []interface{}{1, 2, 3}
			},
			schema: schematest.NewRef("array",
				schematest.WithItems("integer", schematest.WithXml(
					&schema.Xml{
						Name: "item",
					},
				)),
				schematest.WithXml(&schema.Xml{
					Name:    "foo",
					Wrapped: true,
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo><item>1</item><item>2</item><item>3</item></foo>`, s)
			},
		},
		{
			name: "data is not array",
			data: func() interface{} {
				return 4
			},
			schema: schematest.NewRef("array",
				schematest.WithItems("integer"),
				schematest.WithXml(&schema.Xml{
					Name:    "foo",
					Wrapped: true,
				})),
			test: func(t *testing.T, s string, err error) {
				require.EqualError(t, err, "encoding data to 'application/xml' failed: expected array but got: 4")
			},
		},
		{
			name: "array items not defined",
			data: func() interface{} {
				return []interface{}{1, 2, 3}
			},
			schema: schematest.NewRef("array",
				schematest.WithXml(&schema.Xml{
					Name:    "foo",
					Wrapped: true,
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "<foo><foo>1</foo><foo>2</foo><foo>3</foo></foo>", s)
			},
		},
		{
			name: "array with nil item",
			data: func() interface{} {
				return []interface{}{"bar", nil}
			},
			schema: schematest.NewRef("array",
				schematest.WithItems("string", schematest.IsNullable(true)),
				schematest.WithXml(&schema.Xml{
					Name:    "foo",
					Wrapped: true, // without wrapped results in invalid xml when on root
				})),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<foo><foo>bar</foo></foo>`, s)
			},
		},
		{
			"object from sortedmap",
			func() interface{} {
				m := sortedmap.NewLinkedHashMap()
				m.Set("id", 123)
				m.Set("title", "foo")
				m.Set("x", "2023")
				m.Set("author", "bar")
				return m
			},
			schematest.NewRef("object",
				schematest.WithXml(&schema.Xml{Name: "book"}),
				schematest.WithProperty("id", schematest.New("integer", schematest.WithXml(&schema.Xml{Attribute: true}))),
				schematest.WithProperty("title", schematest.New("string",
					schematest.WithXml(&schema.Xml{
						Attribute: true,
						Prefix:    "ns",
					}),
				)),
				schematest.WithProperty("x", schematest.New("integer",
					schematest.WithXml(&schema.Xml{
						Attribute: true,
						Name:      "year",
					}),
				)),
				schematest.WithProperty("author", schematest.New("string"))),
			func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<book id="123" ns:title="foo" year="2023"><author>bar</author></book>`, s)
			},
		},
		{
			"object with map and empty schema",
			func() interface{} {
				var i interface{}
				i = map[string]interface{}{
					"id":     123,
					"title":  "foo",
					"x":      "2023",
					"author": "bar",
					"foo":    nil,
				}
				return []interface{}{i}
			},
			&schema.Ref{Value: &schema.Schema{Xml: &schema.Xml{Name: "root"}}},
			func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Len(t, s, 74)
				require.True(t, strings.HasPrefix(s, "<root>"))
				require.True(t, strings.HasSuffix(s, "</root>"))
				require.Contains(t, s, "<id>123</id>")
				require.Contains(t, s, "<title>foo</title>")
				require.Contains(t, s, "<x>2023</x>")
				require.Contains(t, s, "<author>bar</author>")
			},
		},
		{
			"nil object",
			func() interface{} {
				return nil
			},
			schematest.NewRef("object",
				schematest.IsNullable(true),
				schematest.WithXml(&schema.Xml{Name: "book"}),
				schematest.WithProperty("id", schematest.New("integer", schematest.WithXml(&schema.Xml{Attribute: true}))),
				schematest.WithProperty("title", schematest.New("string",
					schematest.WithXml(&schema.Xml{
						Attribute: true,
						Prefix:    "ns",
					}),
				)),
				schematest.WithProperty("author", schematest.New("string"))),
			func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<book></book>`, s)
			},
		},
		{
			"object with nil properties",
			func() interface{} {
				m := sortedmap.NewLinkedHashMap()
				m.Set("id", 123)
				m.Set("title", nil)
				m.Set("author", nil)
				return m
			},
			schematest.NewRef("object",
				schematest.WithXml(&schema.Xml{Name: "book"}),
				schematest.WithProperty("id", schematest.New("integer", schematest.WithXml(&schema.Xml{Attribute: true}))),
				schematest.WithProperty("title", schematest.New("string", schematest.IsNullable(true),
					schematest.WithXml(&schema.Xml{
						Attribute: true,
						Prefix:    "ns",
					}),
				)),
				schematest.WithProperty("author", schematest.New("string", schematest.IsNullable(true)))),
			func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<book id="123"></book>`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := tc.schema.Marshal(tc.data(), media.ParseContentType("application/xml"))
			tc.test(t, string(b), err)
		})
	}
}
