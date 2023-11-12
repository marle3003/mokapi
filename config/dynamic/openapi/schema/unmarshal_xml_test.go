package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
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
				require.EqualError(t, err, "unmarshal data failed: EOF")
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
			i, err := r.Unmarshal([]byte(tc.xml), media.ParseContentType("application/xml"))
			tc.test(t, i, err)
		})
	}
}
