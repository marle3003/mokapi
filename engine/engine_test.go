package engine

import (
	"fmt"
	"mokapi/test"
	"testing"
	"time"
)

func TestScriptEngine(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", "export default function(){}")
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
	})
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", "")
		test.EqualError(t, "no exported functions in script", err)
		test.Assert(t, len(engine.scripts) == 0, "script length not 0")
	})
}

func TestEvery(t *testing.T) {
	t.Parallel()
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import mokapi from 'mokapi'
			export default function() {
				mokapi.every('1m', function() {});
			}
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")

		test.Assert(t, len(engine.scripts["test.js"].jobs) == 1, "job not defined")
		test.Assert(t, len(engine.cron.Jobs()) == 1, "job not defined")
	})
}

func TestOn(t *testing.T) {
	t.Parallel()
	t.Run("noEvent", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {}
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
		test.Assert(t, len(engine.scripts["test.js"].events["http"]) == 0, "event defined")
	})
	t.Run("noEmptySummary", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					return false
				});
			}
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
		test.Assert(t, len(engine.scripts["test.js"].events["http"]) == 1, "event not defined")

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 0, "summary length not 0")
	})
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					return true
				});
			}
		`)
		test.Ok(t, err)
		test.Assert(t, len(engine.scripts) == 1, "script length not 1")
		test.Assert(t, len(engine.scripts["test.js"].events["http"]) == 1, "event not defined")

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		summary := summaries[0]
		// tags
		test.Assert(t, summary.Tags["name"] == "test.js", "tag name not correct")
		test.Assert(t, summary.Tags["event"] == "http", "tag event not correct")
	})
	t.Run("duration", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					sleep(1000);
					return true
				});
			}
		`)
		test.Ok(t, err)

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		summary := summaries[0]
		test.Assert(t, summary.Duration > 1.0*time.Second, "sleep")
	})
	t.Run("tag name", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {'name': 'foobar'}});
			}
		`)
		test.Ok(t, err)

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		test.Assert(t, summaries[0].Tags["name"] == "foobar", "tag name not correct")
	})
	t.Run("custom tag", func(t *testing.T) {
		t.Parallel()
		engine := New()
		err := engine.AddScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {'foo': 'bar'}});
			}
		`)
		test.Ok(t, err)

		summaries := engine.Run("http")

		test.Assert(t, len(summaries) == 1, "summary length not 1")
		test.Assert(t, summaries[0].Tags["foo"] == "bar", "tag name not correct")
	})
	t.Run("parameter", func(t *testing.T) {
		t.Parallel()

		p := struct {
			Foo string `js:"foo"`
		}{
			"bar",
		}

		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		engine := NewWithLogger(logger)
		err := engine.AddScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on(
					'http', 
					function(p) {
						console.log(p.foo);
					}
				);
			}
		`)
		test.Ok(t, err)

		engine.Run("http", p)

		test.Equals(t, "bar", msg)
	})
}

type testLogger struct {
	info  func(args ...interface{})
	warn  func(args ...interface{})
	error func(args ...interface{})
}

func (tl *testLogger) Info(args ...interface{}) {
	tl.info(args...)
}

func (tl *testLogger) Warn(args ...interface{}) {
	tl.info(args...)
}

func (tl *testLogger) Error(args ...interface{}) {
	tl.info(args...)
}
