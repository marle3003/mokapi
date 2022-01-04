package modules

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"testing"
)

func TestYaml(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("yaml", YamlLoader)
		err := l.DoString(`
local yaml = require("yaml")
result = yaml.parse("")
`)
		require.NoError(t, err)
		result := l.GetGlobal("result").(*lua.LUserData).Value
		m := result.(map[string]interface{})
		require.Len(t, m, 0)
	})

	t.Run("simple", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("yaml", YamlLoader)
		err := l.DoString(`
local yaml = require("yaml")
result = yaml.parse("foo: bar")
`)
		require.NoError(t, err)
		result := l.GetGlobal("result").(*lua.LUserData).Value
		m := result.(map[string]interface{})
		require.Equal(t, map[string]interface{}{"foo": "bar"}, m)
	})
}
