package generator

import (
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
