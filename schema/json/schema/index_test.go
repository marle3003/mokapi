package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestNewIndexData(t *testing.T) {
	idx := schema.NewIndexData(nil)
	require.Nil(t, idx)

	s := &schema.Schema{
		Title:       "foo",
		Description: "bar",
	}

	idx = schema.NewIndexData(s)
	require.Equal(t, "foo", idx.Title)
	require.Equal(t, "bar", idx.Description)
	require.Len(t, idx.Children, 0)

	s = schematest.New("", schematest.WithProperty("foo", &schema.Schema{
		Title:       "foo",
		Description: "bar",
	}))
	idx = schema.NewIndexData(s)
	require.Equal(t, "foo", idx.Children["foo"].Title)
	require.Equal(t, "bar", idx.Children["foo"].Description)
	require.Len(t, idx.Children["foo"].Children, 0)
}
