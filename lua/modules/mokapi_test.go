package modules

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"mokapi/engine/common"
	"testing"
)

func TestMokapi_Every(t *testing.T) {
	t.Run("mokapi.every", func(t *testing.T) {
		var every string
		host := &testHost{
			fnEvery: func(s string, do func(), opt common.JobOptions) (int, error) {
				every = s
				return 0, nil
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
mokapi.every("1m", function() end)
`)

		require.NoError(t, err)
		require.Equal(t, "1m", every)
	})

	t.Run("mokapi.every times", func(t *testing.T) {
		var times int
		host := &testHost{
			fnEvery: func(every string, do func(), opt common.JobOptions) (int, error) {
				times = opt.Times
				return 0, nil
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
mokapi.every("1m", function() end, {times = 3})
`)
		require.NoError(t, err)

		require.Equal(t, 3, times)
	})

	t.Run("mokapi.every tags", func(t *testing.T) {
		var m map[string]string
		host := &testHost{
			fnEvery: func(every string, do func(), opt common.JobOptions) (int, error) {
				m = opt.Tags
				return 0, nil
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
mokapi.every("1m", function() end, {tags = {tag1 = "foo", tag2 = "bar"}})
`)
		require.NoError(t, err)

		require.Equal(t, map[string]string{"tag1": "foo", "tag2": "bar"}, m)
	})

	t.Run("mokapi.every access variable", func(t *testing.T) {
		var fn func()
		host := &testHost{
			fnEvery: func(every string, do func(), opt common.JobOptions) (int, error) {
				fn = do
				return 0, nil
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
local foo = "bar"
mokapi.every("1m", function()
print(foo)
end)
`)
		require.NoError(t, err)

		fn()
	})
}

func TestMokapi_On(t *testing.T) {
	t.Run("mokapi.on event", func(t *testing.T) {
		var event string
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				event = evt
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require "mokapi"
mokapi.on("foo", function() end)
`)

		require.NoError(t, err)
		require.Equal(t, "foo", event)
	})

	t.Run("mokapi.on do returns true", func(t *testing.T) {
		var fn func(args ...interface{}) (bool, error)
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				fn = do
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
mokapi.on("foo", function() return true end)
`)
		require.NoError(t, err)

		b, err := fn()
		require.NoError(t, err)
		require.True(t, b)
	})

	t.Run("mokapi.on do got error", func(t *testing.T) {
		var fn func(args ...interface{}) (bool, error)
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				fn = do
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")

mokapi.on("foo", function()
foo()
return true
end)
`)
		require.NoError(t, err)

		_, err = fn()
		require.Error(t, err)
	})

	t.Run("mokapi.on with parameters", func(t *testing.T) {
		var fn func(args ...interface{}) (bool, error)
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				fn = do
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
mokapi.on("foo", function(request)
request.data = {xy = {1, 2}}
return true
end)
`)
		require.NoError(t, err)

		r := &request{}
		b, err := fn(r)
		require.NoError(t, err)
		require.True(t, b)
	})

	t.Run("mokapi.on tags", func(t *testing.T) {
		var m map[string]string
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				m = tags
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
mokapi.on("foo", function() return true end, {tags = {tag1 = "foo", tag2 = "bar"}})
`)
		require.NoError(t, err)

		require.Equal(t, map[string]string{"tag1": "foo", "tag2": "bar"}, m)
	})

	t.Run("mokapi.on access variable", func(t *testing.T) {
		var fn func(args ...interface{}) (bool, error)
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				fn = do
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
local foo = "bar"
mokapi.on("foo", function(request)
request.data = foo
return true end)
`)
		require.NoError(t, err)

		r := &request{}
		b, err := fn(r)
		require.NoError(t, err)
		require.True(t, b)
		require.Equal(t, "bar", r.Data)
	})

	t.Run("mokapi.on two handlers", func(t *testing.T) {
		var fns []func(args ...interface{}) (bool, error)
		host := &testHost{
			fnOn: func(evt string, do func(args ...interface{}) (bool, error), tags map[string]string) {
				fns = append(fns, do)
			},
		}
		l := lua.NewState(lua.Options{IncludeGoStackTrace: true})
		defer l.Close()

		l.PreloadModule("mokapi", NewMokapi(host).Loader)
		err := l.DoString(`
local mokapi = require("mokapi")
 
mokapi.on("foo", function(request, response)
response.data = "bar"
return true end)
mokapi.on("foo", function(request, response)
response.headers["foo"] = "bar"
return true end)
`)
		require.NoError(t, err)

		r := &request{}
		res := &response{Headers: map[string]interface{}{}}

		for _, fn := range fns {
			b, err := fn(r, res)
			require.NoError(t, err)
			require.True(t, b)
		}

		require.Equal(t, "bar", res.Data)
		require.Equal(t, "bar", res.Headers["foo"])
	})
}

type testHost struct {
	common.Host
	fnInfo  func(s string)
	fnOn    func(event string, do func(args ...interface{}) (bool, error), tags map[string]string)
	fnEvery func(every string, do func(), opt common.JobOptions) (int, error)
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

func (th *testHost) Every(every string, do func(), opt common.JobOptions) (int, error) {
	if th.fnEvery != nil {
		return th.fnEvery(every, do, opt)
	}
	panic("not implemented")
}

type request struct {
	Data interface{}
}

type response struct {
	Data    interface{}
	Headers map[string]interface{}
}
