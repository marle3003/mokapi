package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext_SnapshotRestore(t *testing.T) {
	t.Run("nested level reads the values of its parent", func(t *testing.T) {
		c := newContext()
		c.Values["firstname"] = "Emma"

		c.Snapshot() // root object
		c.Snapshot() // nested object

		require.Equal(t, "Emma", c.Values["firstname"])
	})

	t.Run("values of a nested level are discarded on restore", func(t *testing.T) {
		c := newContext()
		c.Snapshot() // root object
		c.Values["firstname"] = "Emma"

		c.Snapshot() // nested object
		c.Values["city"] = "Bern"
		c.Restore()

		require.Equal(t, map[string]any{"firstname": "Emma"}, c.Values)
	})

	t.Run("sibling levels do not see each other", func(t *testing.T) {
		c := newContext()
		c.Snapshot() // root object

		c.Snapshot()
		c.Values["city"] = "Bern"
		c.Restore()

		c.Snapshot()
		require.NotContains(t, c.Values, "city")
		c.Restore()
	})

	t.Run("deeply nested levels are restored one by one", func(t *testing.T) {
		c := newContext()
		c.Snapshot()
		c.Values["a"] = 1
		c.Snapshot()
		c.Values["b"] = 2
		c.Snapshot()
		c.Values["c"] = 3

		c.Restore()
		require.Equal(t, map[string]any{"a": 1, "b": 2}, c.Values)
		c.Restore()
		require.Equal(t, map[string]any{"a": 1}, c.Values)
	})

	t.Run("restore on an empty stack", func(t *testing.T) {
		c := newContext()
		c.Values["firstname"] = "Emma"

		c.Restore()

		require.Equal(t, map[string]any{"firstname": "Emma"}, c.Values)
	})
}

func TestContext_ParametersAndLevels(t *testing.T) {
	schemaWithNestedId := schematest.New("object",
		schematest.WithProperty("id", schematest.New("integer")),
		schematest.WithProperty("friend", schematest.New("object",
			schematest.WithProperty("id", schematest.New("integer")))))

	t.Run("parameter is used at its shallowest level only", func(t *testing.T) {
		c := NewRequest(nil, schemaWithNestedId, map[string]any{"id": 1}).Context

		v, ok := c.valueAt("id", 1)
		require.True(t, ok)
		require.Equal(t, 1, v)
		_, ok = c.valueAt("id", 2)
		require.False(t, ok)
	})

	t.Run("parameter that no schema declares is used at any level", func(t *testing.T) {
		c := NewRequest(nil, schemaWithNestedId, map[string]any{"other": 1}).Context

		_, ok := c.valueAt("other", 5)
		require.True(t, ok)
	})

	t.Run("value written by a node is used at any level", func(t *testing.T) {
		c := NewRequest(nil, schemaWithNestedId, map[string]any{"id": 1}).Context
		c.Values["firstname"] = "Emma"

		_, ok := c.valueAt("firstname", 3)
		require.True(t, ok)
	})

	t.Run("missing value", func(t *testing.T) {
		c := NewRequest(nil, schemaWithNestedId, map[string]any{"id": 1}).Context

		_, ok := c.valueAt("foo", 1)
		require.False(t, ok)
	})

	t.Run("forget node keeps parameters", func(t *testing.T) {
		c := NewRequest(nil, schemaWithNestedId, map[string]any{"id": 1}).Context
		c.Values["firstname"] = "Emma"

		c.forgetNode(&Node{Name: "id"})
		c.forgetNode(&Node{Name: "firstname"})
		c.forgetNode(nil)

		require.Equal(t, map[string]any{"id": 1}, c.Values)
	})
}
