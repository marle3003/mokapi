package schema_test

import (
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseXML(t *testing.T) {
	testcases := []struct {
		name string
		xml  string
		s    *schema.Schema
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "value expected in attribute but it is located in child",
			xml:  `<root><id>123</id></root>`,
			s: schematest.New("object",
				schematest.WithProperty(
					"id",
					schematest.New("integer", schematest.WithXml(&schema.Xml{Attribute: true})),
				),
			),
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				// id is indeed defined as attribute but in the payload as node
				// so it shouldn't be treated as an additional property
				require.Equal(t, map[string]any{}, v)
			},
		},
		{
			name: "value expected in attribute but it is located in child and additionalProperties not allowed",
			xml:  `<root><id>123</id></root>`,
			s: schematest.New("object",
				schematest.WithProperty(
					"id",
					schematest.New("integer", schematest.WithXml(&schema.Xml{Attribute: true})),
				),
				schematest.WithFreeForm(false),
			),
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "failed to parse XML: property 'id' is expected as XML attribute but received as XML node and additionalProperty is not allowed: /root/id")
			},
		},
		{
			name: "attribute required but value is located in child",
			xml:  `<root><id>123</id></root>`,
			s: schematest.New("object",
				schematest.WithProperty(
					"id",
					schematest.New("integer", schematest.WithXml(&schema.Xml{Attribute: true})),
				),
				schematest.WithRequired("id"),
			),
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "failed to parse XML: required attribute 'id' not found in XML")
			},
		},
		{
			name: "attribute required using space",
			xml:  `<root ns:id="123"></root>`,
			s: schematest.New("object",
				schematest.WithProperty(
					"id",
					schematest.New("integer", schematest.WithXml(&schema.Xml{
						Attribute: true,
						Prefix:    "ns",
					})),
				),
				schematest.WithRequired("id"),
			),
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"id": int64(123)}, v)
			},
		},
		{
			name: "attribute required using space but missing",
			xml:  `<root><foo>bar</foo></root>`,
			s: schematest.New("object",
				schematest.WithProperty(
					"id",
					schematest.New("integer", schematest.WithXml(&schema.Xml{
						Attribute: true,
						Prefix:    "ns",
					})),
				),
				schematest.WithRequired("id"),
			),
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "failed to parse XML: required attribute 'id' not found in XML")
			},
		},
		{
			name: "attribute required empty node",
			xml:  `<root></root>`,
			s: schematest.New("object",
				schematest.WithProperty(
					"id",
					schematest.New("integer", schematest.WithXml(&schema.Xml{
						Attribute: true,
						Prefix:    "ns",
					})),
				),
				schematest.WithRequired("id"),
			),
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/required: required properties are missing: id")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := schema.NewXmlParser(tc.s)
			v, err := p.Parse(tc.xml)
			tc.test(t, v, err)
		})
	}
}

func TestUnmarshalXML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "empty",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(""), schematest.New(""))
				require.NoError(t, err)
				require.Nil(t, v)
			},
		},
		{
			name: "no schema defined",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<book>Harry Potter</book>"), nil)
				require.NoError(t, err)
				require.Equal(t, "Harry Potter", v)
			},
		},
		{
			name: "string",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<book>Harry Potter</book>"), schematest.New("string"))
				require.NoError(t, err)
				require.Equal(t, "Harry Potter", v)
			},
		},
		{
			name: "integer",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<year>2005</year>"), schematest.New("integer"))
				require.NoError(t, err)
				require.Equal(t, int64(2005), v)
			},
		},
		{
			name: "number",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<price>19.90</price>"), schematest.New("number"))
				require.NoError(t, err)
				require.Equal(t, 19.90, v)
			},
		},
		{
			name: "boolean",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<disabled>true</disabled>"), schematest.New("boolean"))
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name: "object with one property",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<book><title>Harry Potter</title></book>"),
					schematest.New("object",
						schematest.WithProperty("title", schematest.New("string"))))
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"title": "Harry Potter"}, v)
			},
		},
		{
			name: "object with two property",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<book><title>Harry Potter</title><author>J K. Rowling</author></book>"),
					schematest.New("object",
						schematest.WithProperty("title", schematest.New("string"))))
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"title": "Harry Potter", "author": "J K. Rowling"}, v)
			},
		},
		{
			name: "object with different XML name",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<book><title>Harry Potter</title></book>"),
					schematest.New("object",
						schematest.WithProperty(
							"name",
							schematest.New("string", schematest.WithXml(&schema.Xml{Name: "title"})),
						),
					))
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Harry Potter"}, v)
			},
		},
		{
			name: "two properties with misleading name",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<root><foo>bar</foo><bar>123</bar></root>"),
					schematest.New("object",
						schematest.WithProperty(
							"foo",
							schematest.New("integer", schematest.WithXml(&schema.Xml{Name: "bar"})),
						),
						schematest.WithProperty(
							"yuh",
							schematest.New("string", schematest.WithXml(&schema.Xml{Name: "foo"})),
						),
					))
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(123), "yuh": "bar"}, v)
			},
		},
		{
			name: "object property as attribute",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<person name="alice"></person>`),
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string", schematest.WithXml(&schema.Xml{Attribute: true})))))
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "alice"}, v)
			},
		},
		{
			name: "object property as attribute and different name",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<person firstname="alice"></person>`),
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string", schematest.WithXml(&schema.Xml{Attribute: true, Name: "firstname"})))))
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "alice"}, v)
			},
		},
		{
			name: "object free-form",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<person name="alice"><age>29</age></person>`), nil)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "alice", "age": "29"}, v)
			},
		},
		{
			name: "string array",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<books><title>Harry Potter</title></books>"),
					schematest.New("array",
						schematest.WithItems("string"),
					),
				)
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Harry Potter"}, v)
			},
		},
		{
			name: "string array with two items",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<books><title>Harry Potter</title><title>Hush</title></books>"),
					schematest.New("array",
						schematest.WithItems("string"),
					),
				)
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Harry Potter", "Hush"}, v)
			},
		},
		{
			name: "array without schema",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<books><title>Harry Potter</title><title>Hush</title></books>"), nil)
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Harry Potter", "Hush"}, v)
			},
		},
		{
			name: "array as property",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<person><children><name>bob</name><name>sarah</name></children></person>"),
					schematest.New("object", schematest.WithProperty("children", schematest.New("array",
						schematest.WithItems("string", schematest.WithXml(&schema.Xml{Name: "name"})),
						schematest.WithXml(&schema.Xml{Wrapped: true}),
					))),
				)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"children": []interface{}{"bob", "sarah"}}, v)
			},
		},
		{
			name: "array as property and wrapped",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader("<root><books> <books>one</books> <books>two</books>  <books>three</books> </books> </root>"),
					schematest.New("array",
						schematest.WithItems("string"),
						schematest.WithXml(&schema.Xml{Wrapped: true, Name: "person"}),
					),
				)
				require.NoError(t, err)
				require.Equal(t, []interface{}{"one", "two", "three"}, v)
			},
		},
		{
			name: "prefix and namespace",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<smp:book xmlns:smp="https://example.com/schema"><smp:id>15</smp:id><smp:title>Harry Potter</smp:title><smp:author>J K. Rowling</smp:author></smp:book>`),
					schematest.New("object"),
				)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"id": "15", "author": "J K. Rowling", "title": "Harry Potter", "smp": "https://example.com/schema"}, v)
			},
		},
		{
			name: "prefix and namespace with property schemas",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<smp:book xmlns:smp="https://example.com/schema"><smp:id>15</smp:id><smp:title>Harry Potter</smp:title><smp:author>J K. Rowling</smp:author></smp:book>`),
					schematest.New("object",
						schematest.WithProperty("author", schematest.New("string")),
						schematest.WithProperty("id", schematest.New("integer")),
						schematest.WithProperty("title", schematest.New("string")),
					),
				)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"id": int64(15), "author": "J K. Rowling", "title": "Harry Potter", "smp": "https://example.com/schema"}, v)
			},
		},
		{
			name: "prefix and namespace prefix not match",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<smp:book xmlns:smp="https://example.com/schema"><smp:id>15</smp:id><smp:title>Harry Potter</smp:title><smp:author>J K. Rowling</smp:author></smp:book>`),
					schematest.New("object",
						schematest.WithProperty("author", schematest.New("string")),
						schematest.WithProperty("id", schematest.New("integer", schematest.WithXml(&schema.Xml{Prefix: "foo"}))),
						schematest.WithProperty("title", schematest.New("string")),
					),
				)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"id": "15", "author": "J K. Rowling", "title": "Harry Potter", "smp": "https://example.com/schema"}, v)
			},
		},
		{
			name: "books example from swagger",
			test: func(t *testing.T) {
				v, err := schema.UnmarshalXML(strings.NewReader(`<root><foo><books>one</books><books>two</books></foo></root>`),
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("object",
							schematest.WithProperty("books",
								schematest.New("array", schematest.WithItems("string"))),
						)),
					),
				)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"foo": map[string]any{"books": []any{"one", "two"}}}, v)
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

func TestUnmarshalXML_Old(t *testing.T) {
	testcases := []struct {
		name   string
		xml    string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "free object",
			xml:    "<book><id>0</id><title>foo</title><author>bar</author></book>",
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"id": "0", "title": "foo", "author": "bar"}, i)
			},
		},
		{
			name:   "free array",
			xml:    "<books><book>one</book><book>two</book></books>",
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"one", "two"}, i)
			},
		},
		{
			name: "wrapped array",
			xml:  "<root><books><books>one</books><books>two</books></books></root>",
			schema: schematest.New("array", schematest.WithItems("string"), schematest.WithXml(&schema.Xml{
				Wrapped: true,
			})),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"one", "two"}, i)
			},
		},
		{
			name: "simple",
			xml:  "<book><id>0</id><title>foo</title><author>bar</author></book>",
			schema: schematest.New("object",
				schematest.WithXml(&schema.Xml{Name: "book"}),
				schematest.WithProperty("id", schematest.New("integer")),
				schematest.WithProperty("title", schematest.New("string")),
				schematest.WithProperty("author", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"id": int64(0), "title": "foo", "author": "bar"}, i)
			},
		},
		{
			name: "xml name",
			xml:  "<Book><Id>0</Id><Title>foo</Title><Author>bar</Author></Book>",
			schema: schematest.New("object",
				schematest.WithXml(&schema.Xml{Name: "Book"}),
				schematest.WithProperty("id", schematest.New("integer", schematest.WithXml(&schema.Xml{Name: "Id"}))),
				schematest.WithProperty("title", schematest.New("string", schematest.WithXml(&schema.Xml{Name: "Title"}))),
				schematest.WithProperty("author", schematest.New("string", schematest.WithXml(&schema.Xml{Name: "Author"})))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"id": int64(0), "title": "foo", "author": "bar"}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v, err := schema.UnmarshalXML(strings.NewReader(tc.xml), tc.schema)
			tc.test(t, v, err)
		})
	}
}

func TestUnmarshalXML_NoSchema(t *testing.T) {
	testcases := []struct {
		name string
		xml  string
		test func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "string",
			xml:  "<root>foo</root>",
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", i)
			},
		},
		{
			name: "same name",
			xml:  "<root><name>alice</name><name>carol</name></root>",
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"alice", "carol"}, i)
			},
		},
		{
			name: "mixed with same name",
			xml:  "<root><foo>bar</foo><name>alice</name><name>carol</name></root>",
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar", "name": []interface{}{"alice", "carol"}}, i)
			},
		},
		{
			name: "mixed object property and array list containing multiple entries",
			xml:  "<root><foo>bar</foo><name>alice</name><name>carol</name><name>sarah</name></root>",
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar", "name": []interface{}{"alice", "carol", "sarah"}}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v, err := schema.UnmarshalXML(strings.NewReader(tc.xml), nil)
			tc.test(t, v, err)
		})
	}
}
