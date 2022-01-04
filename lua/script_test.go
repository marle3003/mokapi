package lua

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"testing"
)

func TestScript(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		host := &testHost{}
		s, err := New("foo.lua", "", host)
		require.NoError(t, err)
		defer s.Close()
		err = s.Run()
		require.NoError(t, err)
	})

	t.Run("syntax error", func(t *testing.T) {
		host := &testHost{}
		s, err := New("foo.lua", "foo()", host)
		require.NoError(t, err)
		defer s.Close()
		err = s.Run()
		require.Error(t, err)
	})

	t.Run("log", func(t *testing.T) {
		var log string
		host := &testHost{
			fnInfo: func(s string) {
				log = s
			},
		}
		s, err := New("foo.lua", "log:info(\"foobar\")", host)
		require.NoError(t, err)
		defer s.Close()
		err = s.Run()
		require.NoError(t, err)
		require.Equal(t, "foobar", log)
	})
}

func TestMokapi_On(t *testing.T) {
	t.Run("mokapi:on", func(t *testing.T) {
		called := false
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				called = true
			},
		}
		s, err := New("foo.lua", `
local mokapi = require("mokapi")
mokapi:on("foo", function() end)
`, host)

		require.NoError(t, err)
		defer s.Close()
		err = s.Run()
		require.NoError(t, err)
		require.True(t, called)
	})
}

func TestYaml(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {
		host := &testHost{}
		s, err := New("foo.lua", `
local yaml = require("yaml")
yaml.parse("")
`, host)

		require.NoError(t, err)
		defer s.Close()
		err = s.Run()
		require.NoError(t, err)
	})
}

func TestMustache(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {
		host := &testHost{}
		s, err := New("foo.lua", `
local mustache = require("mustache")
mustache.render("", {})
`, host)

		require.NoError(t, err)
		defer s.Close()
		err = s.Run()
		require.NoError(t, err)
	})
}

type testHost struct {
	common.Host
	fnInfo func(s string)
	fnOn   func(event string, do func(args ...interface{}) (bool, error), tags map[string]string)
}

func (th *testHost) Info(args ...interface{}) {
	if th.fnInfo != nil {
		th.fnInfo(args[0].(string))
	}
}

func (th *testHost) On(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
	if th.fnOn != nil {
		th.fnOn(event, do, tags)
	}
}
