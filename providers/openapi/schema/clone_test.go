package schema_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// parseCyclic returns a schema whose resolved $ref graph is cyclic:
// node.properties.child refers to node.
func parseCyclic(t *testing.T) *schema.Schema {
	t.Helper()
	var s *schema.Schema
	err := yaml.Unmarshal([]byte(`
$defs:
  node:
    type: object
    title: node
    properties:
      child:
        $ref: '#/$defs/node'
$ref: '#/$defs/node'
`), &s)
	require.NoError(t, err)
	require.NoError(t, s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{}))

	node := s.Sub
	require.Same(t, node, node.Properties.Get("child").Sub, "the graph must be cyclic")
	return s
}

func TestSchema_Clone(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var s *schema.Schema
		require.Nil(t, s.Clone())
		var ss *schema.Schemas
		require.Nil(t, ss.Clone())
	})

	t.Run("copy does not share with the original", func(t *testing.T) {
		min := 1
		s := &schema.Schema{
			Type:       jsonSchema.Types{"object"},
			Required:   []string{"id"},
			MinLength:  &min,
			Properties: &schema.Schemas{},
		}
		s.Properties.Set("id", &schema.Schema{Type: jsonSchema.Types{"string"}})

		c := s.Clone()
		c.Type[0] = "array"
		c.Required[0] = "name"
		*c.MinLength = 5
		c.Properties.Get("id").Type[0] = "integer"

		require.Equal(t, jsonSchema.Types{"object"}, s.Type)
		require.Equal(t, []string{"id"}, s.Required)
		require.Equal(t, 1, *s.MinLength)
		require.Equal(t, jsonSchema.Types{"string"}, s.Properties.Get("id").Type)
	})

	t.Run("cyclic schema", func(t *testing.T) {
		s := parseCyclic(t)

		c := s.Clone()

		node := c.Sub
		require.NotSame(t, s.Sub, node)
		require.Equal(t, "node", node.Title)
		require.Same(t, node, node.Properties.Get("child").Sub, "the copy has the cycle of the original")

		node.Title = "changed"
		require.Equal(t, "node", s.Sub.Title)
	})
}
