package compiler

import (
	"mokapi/test"
	"testing"
)

func TestScript(t *testing.T) {
	t.Parallel()
	c, err := New()
	test.Ok(t, err)
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		_, err := c.Compile("", "")
		test.Ok(t, err)
	})
	t.Run("code", func(t *testing.T) {
		t.Parallel()
		_, err := c.Compile("", "function test() {}")
		test.Ok(t, err)
	})
	t.Run("syntax error", func(t *testing.T) {
		t.Parallel()
		_, err := c.Compile("", "function test()")
		test.Equals(t, "SyntaxError: /test.js: Unexpected token, expected \"{\" (1:15)\n\n> 1 | function test()\n    |                ^ at dispatchException (mokapi/babel.min.js:2:728343(7))", err.Error())
	})
}
