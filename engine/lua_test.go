package engine

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/runtime"
	"net/url"
	"testing"
)

func TestEngine_AddScript(t *testing.T) {
	engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
	src := `
			local mokapi = require "mokapi"
			mokapi.every("1m", function() end);
`
	err := engine.AddScript(newScript("test.lua", src))
	require.NoError(t, err)
	err = engine.AddScript(newScript("test.lua", src))
	require.NoError(t, err)

	require.Len(t, engine.cron.Jobs(), 1, "only one job should exists")
}

func TestLuaScriptEngine(t *testing.T) {
	t.Parallel()
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", ""))
		require.NoError(t, err)
	})
	t.Run("print", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `print("Hello World")`))
		require.NoError(t, err)
	})
}

func TestLuaEvery(t *testing.T) {
	t.Parallel()
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			id = mokapi.every("1m", function() end);
		`))
		require.NoError(t, err)
		require.Len(t, engine.scripts, 1, "script length not 1")

		require.Len(t, engine.scripts["test.lua"].jobs, 1, "job not defined")
		require.Len(t, engine.cron.Jobs(), 1, "job not defined")
	})
}

func TestLuaOn(t *testing.T) {
	t.Parallel()
	t.Run("noEvent", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
		`))
		require.NoError(t, err)
		require.Len(t, engine.scripts, 0, "script should be closed because no event and no jobs")
	})
	t.Run("withoutSummary", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi.on(
				'http',
				function()
					return false
				end
			);
		`))
		require.NoError(t, err)
		require.Len(t, engine.scripts, 1, "script length not 1")
		require.Len(t, engine.scripts["test.lua"].events["http"], 1, "event not defined")

		summaries := engine.Run("http")

		require.Len(t, summaries, 0, "summary length not 0")
	})
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi.on(
				'http',
				function()
					return true	
				end
			);
		`))
		require.NoError(t, err)
		require.Len(t, engine.scripts, 1, "script length not 1")
		require.Len(t, engine.scripts["test.lua"].events["http"], 1, "event not defined")

		summaries := engine.Run("http")

		require.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		// tags
		require.Equal(t, "test.lua", summary.Tags["name"], "tag name not correct")
		require.Equal(t, "http", summary.Tags["event"], "tag event not correct")
	})
	t.Run("duration", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi.on(
				'http',
				function()
					sleep(1000);
					return true	
				end
			);
		`))
		require.NoError(t, err)

		summaries := engine.Run("http")

		require.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		require.True(t, summary.Duration >= 1.0, "sleep")
	})
	t.Run("tag name", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi.on(
				'http',
				function()
					return true	
				end,
				{tags = {name = 'foobar'}}
			);
		`))
		require.NoError(t, err)

		summaries := engine.Run("http")

		require.Len(t, summaries, 1, "summary length not 1")
		require.Equal(t, "foobar", summaries[0].Tags["name"], "tag name not correct")
	})
	t.Run("custom tag", func(t *testing.T) {
		t.Parallel()
		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi.on(
				'http',
				function()
					return true	
				end,
				{tags = {foo = 'bar'}}
			);
		`))
		require.NoError(t, err)

		summaries := engine.Run("http")

		require.Len(t, summaries, 1, "summary length not 1")
		require.Equal(t, "bar", summaries[0].Tags["foo"], "tag name not correct")
	})
	t.Run("parameter", func(t *testing.T) {
		t.Parallel()

		p := struct {
			Foo string
		}{
			"bar",
		}

		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
		engine.logger = logger
		err := engine.AddScript(newScript("test.lua", `
			local mokapi = require "mokapi"
			local log = require "log"
			mokapi.on(
				'http',
				function(p)
					log.info(p.foo)
					return true
				end
			);
		`))
		require.NoError(t, err)

		engine.Run("http", p)

		require.Equal(t, "bar", msg)
	})
}

func TestLuaOpen(t *testing.T) {
	t.Parallel()
	t.Run("fileExists", func(t *testing.T) {
		t.Parallel()
		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
			return &dynamic.Config{
				Info: dynamic.ConfigInfo{Url: u},
				Raw:  []byte("foobar"),
			}, nil
		})

		engine := New(reader, runtime.New(), static.JsConfig{})
		engine.logger = logger
		err := engine.AddScript(newScript("./test.lua", `
			local file = open('test.txt')
			local log = require "log"
			log.info(file)
		`))
		require.NoError(t, err)
		require.Equal(t, "foobar", msg)
	})
	t.Run("fileNotExists", func(t *testing.T) {
		t.Parallel()

		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
			return nil, errors.New("file not found")
		})

		engine := New(reader, runtime.New(), static.JsConfig{})
		engine.logger = logger
		err := engine.AddScript(newScript("./test.lua", `
			local file, err = open('test.txt')
			local log = require "log"
			log.info(err)
		`))
		require.NoError(t, err)
		require.Equal(t, "file not found", msg)
	})
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
