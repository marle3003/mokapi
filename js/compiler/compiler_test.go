package compiler

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScript(t *testing.T) {
	t.Parallel()
	c, err := New()
	require.NoError(t, err)
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		_, err := c.Compile("", "")
		require.NoError(t, err)
	})
	t.Run("code", func(t *testing.T) {
		t.Parallel()
		_, err := c.Compile("", "function test() {}")
		require.NoError(t, err)
	})
	t.Run("syntax error", func(t *testing.T) {
		t.Parallel()
		_, err := c.Compile("", "function test()")
		require.Equal(t, "SyntaxError: /babel.js: Unexpected token, expected \"{\" (1:15)\n\n> 1 | function test()\n    |                ^ at dispatchException (mokapi/babel.min.js:2:728337(7))", err.Error())
	})
}
