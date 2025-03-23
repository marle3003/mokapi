package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"strings"
	"testing"
)

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
				require.Equal(t, 2005, v)
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
						schematest.WithItems("string"),
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
				require.Equal(t, map[string]interface{}{"id": 15, "author": "J K. Rowling", "title": "Harry Potter", "smp": "https://example.com/schema"}, v)
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
			schema: schematest.New("object", schematest.WithProperty("books",
				schematest.New("array", schematest.WithItems("string"), schematest.WithXml(&schema.Xml{
					Wrapped: true,
				})))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"books": []interface{}{"one", "two"}}, i)
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
				require.Equal(t, map[string]interface{}{"id": 0, "title": "foo", "author": "bar"}, i)
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
				require.Equal(t, map[string]interface{}{"id": 0, "title": "foo", "author": "bar"}, i)
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
