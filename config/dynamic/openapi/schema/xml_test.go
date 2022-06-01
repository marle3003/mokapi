package schema_test

import (
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"mokapi/sortedmap"
	"testing"
)

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
				require.Equal(t, err, io.EOF)
			},
		},
		{
			"free object",
			"<book><id>0</id><title>foo</title><author>bar</author></book>",
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				m, ok := i.(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, "0", m["id"])
				require.Equal(t, "foo", m["title"])
				require.Equal(t, "bar", m["author"])
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
				m, ok := i.(*sortedmap.LinkedHashMap)
				require.True(t, ok)
				require.Equal(t, int64(0), m.Get("id"))
				require.Equal(t, "foo", m.Get("title"))
				require.Equal(t, "bar", m.Get("author"))
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

func TestEncode_Xml(t *testing.T) {
	testcases := []struct {
		name   string
		data   func() interface{}
		schema *schema.Schema
		f      func(t *testing.T, s string, err error)
	}{
		{
			"nil",
			func() interface{} {
				return nil
			},
			schematest.New("object"),
			func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "<root></root>", s)
			},
		},
		{
			"simple",
			func() interface{} {
				m := sortedmap.NewLinkedHashMap()
				m.Set("id", 0)
				m.Set("title", "foo")
				m.Set("author", "bar")
				return m
			},
			schematest.New("object",
				schematest.WithXml(&schema.Xml{Name: "book"}),
				schematest.WithProperty("id", schematest.New("integer")),
				schematest.WithProperty("title", schematest.New("string")),
				schematest.WithProperty("author", schematest.New("string"))),
			func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "<book><id>0</id><title>foo</title><author>bar</author></book>", s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &schema.Ref{Value: tc.schema, Reference: ref.Reference{Ref: "#/root"}}
			b, err := r.Marshal(tc.data(), media.ParseContentType("application/xml"))
			tc.f(t, string(b), err)
		})
	}
}
