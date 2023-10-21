package schema_test

import (
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"reflect"
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
				require.Equal(t, io.EOF, err)
			},
		},
		{
			name:   "free object",
			xml:    "<book><id>0</id><title>foo</title><author>bar</author></book>",
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.True(t, hasField(i, "Id", "0"))
				require.True(t, hasField(i, "Title", "foo"))
				require.True(t, hasField(i, "Author", "bar"))
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
				require.Equal(t, &struct {
					Id     int64  `json:"id"`
					Title  string `json:"title"`
					Author string `json:"author"`
				}{Id: 0, Title: "foo", Author: "bar"}, i)
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
				require.Equal(t, &struct {
					Id     int64  `json:"id"`
					Title  string `json:"title"`
					Author string `json:"author"`
				}{Id: 0, Title: "foo", Author: "bar"}, i)
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

func hasField(i interface{}, field string, value interface{}) bool {
	v := reflect.ValueOf(i).Elem()
	f := v.FieldByName(field)
	if !f.IsValid() {
		return false
	}
	return f.Interface() == value
}
