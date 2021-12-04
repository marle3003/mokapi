package engine

import (
	"errors"
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/test"
	"testing"
	"time"
)

func TestLuaScriptEngine(t *testing.T) {
	t.Parallel()
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", "")
		test.Ok(t, err)
	})
	t.Run("print", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `print("Hello World")`)
		test.Ok(t, err)
	})
}

func TestLuaEvery(t *testing.T) {
	t.Parallel()
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:every("1m", function() end);
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")

		test.Assert(t, len(engine.scripts["test.lua"].jobs) == 1, "job not defined")
		test.Assert(t, len(engine.cron.Jobs()) == 1, "job not defined")
	})
	t.Run("simple2", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi.every(mokapi, "1m", function() end);
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")

		test.Assert(t, len(engine.scripts["test.lua"].jobs) == 1, "job not defined")
		test.Assert(t, len(engine.cron.Jobs()) == 1, "job not defined")
	})
}

func TestLuaOn(t *testing.T) {
	t.Parallel()
	t.Run("noEvent", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
		test.Assert(t, len(engine.scripts["test.lua"].events["http"]) == 0, "event defined")
	})
	t.Run("withoutSummary", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:on(
				'http',
				function()
					return false
				end
			);
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
		test.Assert(t, len(engine.scripts["test.lua"].events["http"]) == 1, "event not defined")

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 0, "summary length not 0")
	})
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:on(
				'http',
				function()
					return true	
				end
			);
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
		test.Assert(t, len(engine.scripts["test.lua"].events["http"]) == 1, "event not defined")

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		summary := summaries[0]
		// tags
		test.Assert(t, summary.Tags["name"] == "test.lua", "tag name not correct")
		test.Assert(t, summary.Tags["event"] == "http", "tag event not correct")
	})
	t.Run("duration", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:on(
				'http',
				function()
					sleep(1000);
					return true	
				end
			);
		`)
		test.Ok(t, err)

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		summary := summaries[0]
		test.Assert(t, summary.Duration >= 1.0*time.Second, "sleep")
	})
	t.Run("tag name", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:on(
				'http',
				function()
					return true	
				end,
				{tags = {name = 'foobar'}}
			);
		`)
		test.Ok(t, err)

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		test.Assert(t, summaries[0].Tags["name"] == "foobar", "tag name not correct")
	})
	t.Run("custom tag", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader)
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:on(
				'http',
				function()
					return true	
				end,
				{tags = {foo = 'bar'}}
			);
		`)
		test.Ok(t, err)

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		test.Assert(t, summaries[0].Tags["foo"] == "bar", "tag name not correct")
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

		engine := New(emptyReader)
		engine.logger = logger
		err := engine.AddScript("test.lua", `
			local mokapi = require "mokapi"
			mokapi:on(
				'http',
				function(p)
					log:info(p.foo)
					return true
				end
			);
		`)
		test.Ok(t, err)

		engine.Run("http", p)

		test.Equals(t, "bar", msg)
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

		reader := &testReader{readFunc: func(file *common.File) error {
			file.Data = "foobar"
			return nil
		}}

		engine := New(reader)
		engine.logger = logger
		err := engine.AddScript("./test.lua", `
			local file = open('test.txt')
			log:info(file)
		`)
		test.Ok(t, err)
		test.Equals(t, "foobar", msg)
	})
	t.Run("fileNotExists", func(t *testing.T) {
		t.Parallel()

		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		reader := &testReader{readFunc: func(file *common.File) error {
			return errors.New("file not found")
		}}

		engine := New(reader)
		engine.logger = logger
		err := engine.AddScript("./test.lua", `
			local file, err = open('test.txt')
			log:info(err)
		`)
		test.Ok(t, err)
		test.Equals(t, "file not found", msg)
	})
}
