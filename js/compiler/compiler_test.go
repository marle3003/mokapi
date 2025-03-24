package compiler

import (
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScript(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, c *Compiler)
	}{
		{
			name: "blank",
			test: func(t *testing.T, c *Compiler) {
				_, err := c.Compile("", "")
				require.NoError(t, err)
			},
		},
		{
			name: "syntax error",
			test: func(t *testing.T, c *Compiler) {
				_, err := c.Compile("test.js", "function test()")
				require.EqualError(t, err, "script error: expected \"{\" but found end of file: test.js:1:15")
			},
		},
		{
			name: "es5 function",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", "function test() { return 'foo'}")
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("test()")
				require.NoError(t, err)
				require.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "es6 function",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", "const test = () => { return 'foo'}")
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("test()")
				require.NoError(t, err)
				require.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "Callback",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", `
let result;
const p = Promise.resolve(123);

p.then((value) => {
  result = value
});
`)
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("result")
				require.NoError(t, err)
				require.Equal(t, int64(123), v.Export())
			},
		},
		{
			name: "es8 padStart()",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", `let result = '5'.padStart(4, '0');`)
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("result")
				require.NoError(t, err)
				require.Equal(t, "0005", v.Export())
			},
		},
		{
			name: "es10 trimStart()",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", `let result = '     foo      '.trimStart()`)
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("result")
				require.NoError(t, err)
				require.Equal(t, "foo      ", v.Export())
			},
		},
		{
			name: "es11 optional chaining",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", `const user1 = { address: {number: 12 }}; const user2 = { }; const result = { addr1: user1.address?.number, addr2: user2.address?.number}`)
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("result")
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"addr1": int64(12), "addr2": nil}, v.Export())
			},
		},
		{
			name: "es12 logical nullish assignment",
			test: func(t *testing.T, c *Compiler) {
				prg, err := c.Compile("", `const person1 = { name: 'John'}; person1.name ??= 'Sarah'; const person2 = {}; person2.name ??= 'Sarah'; const result = { p1: person1.name, p2: person2.name }`)
				require.NoError(t, err)
				vm := goja.New()
				_, err = vm.RunProgram(prg)
				require.NoError(t, err)
				v, err := vm.RunString("result")
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"p1": "John", "p2": "Sarah"}, v.Export())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := New()
			require.NoError(t, err)

			tc.test(t, c)
		})
	}
}
