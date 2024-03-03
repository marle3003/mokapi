package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/decoding"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func TestParse_Xml(t *testing.T) {
	testcases := []struct {
		name   string
		xml    string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "empty",
			xml:    "",
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "EOF")
			},
		},
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

			r := &schema.Ref{Value: tc.schema}
			v, err := decoding.Decode([]byte(tc.xml), media.ParseContentType("application/xml"), nil)
			if err != nil {
				tc.test(t, v, err)
				return
			}

			p := schema.Parser{ConvertStringToNumber: true, Xml: true}
			i, err := p.Parse(v, r)
			tc.test(t, i, err)
		})
	}
}
