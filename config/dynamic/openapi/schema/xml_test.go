package schema_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"reflect"
	"testing"
)

func TestRef_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		test func(t *testing.T, r *schema.Xml)
	}{
		{
			name: "xml",
			s: `
  wrapped: true
  name: foo
  attribute: true
  prefix: bar
  namespace: ns1
`,
			test: func(t *testing.T, x *schema.Xml) {
				require.Equal(t, &schema.Xml{
					Wrapped:   true,
					Name:      "foo",
					Attribute: true,
					Prefix:    "bar",
					Namespace: "ns1",
				}, x)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			x := &schema.Xml{}
			err := yaml.Unmarshal([]byte(tc.s), &x)
			require.NoError(t, err)
			tc.test(t, x)
		})
	}
}

func TestParse_Xml(t *testing.T) {
	testcases := []struct {
		name   string
		xml    string
		schema *schema.Schema
		f      func(t *testing.T, i interface{}, err error)
	}{
		{
			"empty",
			"",
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.Equal(t, io.EOF, err)
			},
		},
		{
			"free object",
			"<book><id>0</id><title>foo</title><author>bar</author></book>",
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.True(t, hasField(i, "Id", "0"))
				require.True(t, hasField(i, "Title", "foo"))
				require.True(t, hasField(i, "Author", "bar"))
			},
		},
		{
			"free array",
			"<books><book>one</book><book>two</book></books>",
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"one", "two"}, i)
			},
		},
		{
			"simple",
			"<book><id>0</id><title>foo</title><author>bar</author></book>",
			schematest.New("object",
				schematest.WithXml(&schema.Xml{Name: "book"}),
				schematest.WithProperty("id", schematest.New("integer")),
				schematest.WithProperty("title", schematest.New("string")),
				schematest.WithProperty("author", schematest.New("string"))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Id     int64  `json:"id"`
					Title  string `json:"title"`
					Author string `json:"author"`
				}{Id: 0, Title: "foo", Author: "bar"}, i)
			},
		},
		{
			"xml name",
			"<Book><Id>0</Id><Title>foo</Title><Author>bar</Author></Book>",
			schematest.New("object",
				schematest.WithXml(&schema.Xml{Name: "Book"}),
				schematest.WithProperty("id", schematest.New("integer", schematest.WithXml(&schema.Xml{Name: "Id"}))),
				schematest.WithProperty("title", schematest.New("string", schematest.WithXml(&schema.Xml{Name: "Title"}))),
				schematest.WithProperty("author", schematest.New("string", schematest.WithXml(&schema.Xml{Name: "Author"})))),
			func(t *testing.T, i interface{}, err error) {
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
			i, err := schema.Parse([]byte(tc.xml), media.ParseContentType("application/xml"), &schema.Ref{Value: tc.schema})
			tc.f(t, i, err)
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
