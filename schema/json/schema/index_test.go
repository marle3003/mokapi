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

func TestNewIndexData_SelfReference(t *testing.T) {
	// a schema that is already on the current path is indexed as a leaf
	node := schematest.New("object", func(s *schema.Schema) { s.Title = "Node" })
	node.Properties = &schema.Schemas{}
	node.Properties.Set("name", &schema.Schema{Title: "name"})
	node.Properties.Set("parent", node)

	idx := schema.NewIndexData(node)

	require.Equal(t, "Node", idx.Title)
	require.Equal(t, "name", idx.Children["name"].Title)
	parent := idx.Children["parent"]
	require.NotNil(t, parent)
	require.Equal(t, "Node", parent.Title)
	require.Len(t, parent.Children, 0, "cycle is cut: schema on the current path has no children")
}

func TestNewIndexData_IndirectCycle(t *testing.T) {
	a := &schema.Schema{Title: "A", Properties: &schema.Schemas{}}
	b := &schema.Schema{Title: "B", Properties: &schema.Schemas{}}
	a.Properties.Set("b", b)
	b.Properties.Set("a", a)

	idx := schema.NewIndexData(a)

	require.Equal(t, "B", idx.Children["b"].Title)
	require.Equal(t, "A", idx.Children["b"].Children["a"].Title)
	require.Len(t, idx.Children["b"].Children["a"].Children, 0)
}

func TestNewIndexData_SharedSchemaInSiblingBranches(t *testing.T) {
	// reuse without a cycle must still be fully indexed
	address := schematest.New("object", schematest.WithProperty("street", &schema.Schema{Title: "street"}))
	person := schematest.New("object",
		schematest.WithProperty("home", address),
		schematest.WithProperty("work", address),
	)

	idx := schema.NewIndexData(person)

	require.Equal(t, "street", idx.Children["home"].Children["street"].Title)
	require.Equal(t, "street", idx.Children["work"].Children["street"].Title)
}

func TestNewIndexData_NilProperty(t *testing.T) {
	s := &schema.Schema{Properties: &schema.Schemas{}}
	s.Properties.Set("foo", nil)

	idx := schema.NewIndexData(s)

	require.Nil(t, idx.Children["foo"])
}

func TestNewIndexData_DeepChainIsNotTruncated(t *testing.T) {
	leaf := &schema.Schema{Title: "leaf"}
	mid := &schema.Schema{Title: "mid", Properties: &schema.Schemas{}}
	mid.Properties.Set("leaf", leaf)
	root := &schema.Schema{Title: "root", Properties: &schema.Schemas{}}
	root.Properties.Set("mid", mid)

	idx := schema.NewIndexData(root)

	require.Equal(t, "leaf", idx.Children["mid"].Children["leaf"].Title)
}
