package modules

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"testing"
)

func TestMustache(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("mustache", MustacheLoader)
		err := l.DoString(`
local mustache = require("mustache")
result = mustache.render("", {})
`)
		require.NoError(t, err)
		result := l.GetGlobal("result")
		require.Equal(t, "", result.String())
	})

	t.Run("plain", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("mustache", MustacheLoader)

		err := l.DoString(`
local mustache = require("mustache")
result = mustache.render("foo", {})
`)
		require.NoError(t, err)
		result := l.GetGlobal("result")
		require.Equal(t, "foo", result.String())
	})

	t.Run("simple", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("mustache", MustacheLoader)

		err := l.DoString(`
local mustache = require("mustache")
result = mustache.render("foo{{bar}}", {bar = "rab"})
`)
		require.NoError(t, err)
		result := l.GetGlobal("result")
		require.Equal(t, "foorab", result.String())
	})

	t.Run("nested", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("mustache", MustacheLoader)

		err := l.DoString(`
local mustache = require("mustache")
result = mustache.render("foo{{test.bar}}", {test = {bar = "rab"}})
`)
		require.NoError(t, err)
		result := l.GetGlobal("result")
		require.Equal(t, "foorab", result.String())
	})

	t.Run("error", func(t *testing.T) {
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("mustache", MustacheLoader)

		err := l.DoString(`
local mustache = require("mustache")
result, err = mustache.render("foo{{test.bar}}", {})
`)
		require.NoError(t, err)
		result := l.GetGlobal("result")
		require.Equal(t, lua.LNil, result)

		vErr := l.GetGlobal("err")
		require.Equal(t, `undefined field "test"`, vErr.String())
	})
}
