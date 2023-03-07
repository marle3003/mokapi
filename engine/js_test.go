package engine

import (
	"errors"
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/script"
	"mokapi/runtime"
	"net/url"
	"strings"
	"testing"
)

type testReader struct {
	readFunc func(cfg *common.Config) error
}

func (tr *testReader) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	file := common.NewConfig(u, opts...)
	if err := tr.readFunc(file); err != nil {
		return file, err
	}
	if p, ok := file.Data.(common.Parser); ok {
		return file, p.Parse(file, tr)
	}
	return file, nil
}

func (tr *testReader) Close() {}

var emptyReader = &testReader{readFunc: func(cfg *common.Config) error {
	return nil
}}

func TestJsScriptEngine(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", "export default function(){}"))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 0, "no events and jobs, script should be closed")
	})
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", ""))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 0, "no events and jobs, script should be closed")
	})
}

func TestJsEvery(t *testing.T) {
	t.Parallel()
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import mokapi from 'mokapi'
			export default function() {
				mokapi.every('1m', function() {});
			}
		`))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 1, "script length not 1")

		r.Len(t, engine.scripts["test.js"].jobs, 1, "job not defined")
		r.Len(t, engine.cron.Jobs(), 1, "job not defined")
	})
}

func TestJsOn(t *testing.T) {
	t.Parallel()
	t.Run("noEvent", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {}
		`))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 0, "script length not 1")
	})
	t.Run("withoutSummary", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					return false
				});
			}
		`))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 1, "script length not 1")
		r.Len(t, engine.scripts["test.js"].events["http"], 1, "event not defined")

		summaries := engine.Run("http")

		r.Len(t, summaries, 0, "summary length not 0")
	})
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function(request, response) {
					return true
				});
			}
		`))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 1, "script length not 1")
		r.Len(t, engine.scripts["test.js"].events["http"], 1, "event not defined")

		summaries := engine.Run("http", &struct{}{}, &struct{}{})

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		// tags
		r.Equal(t, "test.js", summary.Tags["name"], "tag name not correct")
		r.Equal(t, "http", summary.Tags["event"], "tag event not correct")
	})
	t.Run("duration", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					sleep(1000);
					return true
				});
			}
		`))
		r.NoError(t, err)

		summaries := engine.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		r.GreaterOrEqual(t, summary.Duration, int64(1000), "sleep")
	})
	t.Run("tag name", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {'name': 'foobar'}});
			}
		`))
		r.NoError(t, err)

		summaries := engine.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		r.Equal(t, "foobar", summaries[0].Tags["name"], "tag name not correct")
	})
	t.Run("custom tag", func(t *testing.T) {
		t.Parallel()
		engine := New(emptyReader, runtime.New())
		err := engine.AddScript(newScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {'foo': 'bar'}});
			}
		`))
		r.NoError(t, err)

		summaries := engine.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		r.Equal(t, "bar", summaries[0].Tags["foo"], "tag name not correct")
	})
	t.Run("parameter", func(t *testing.T) {
		t.Parallel()

		p := struct {
			Foo string `json:"foo"`
		}{
			"bar",
		}

		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		engine := New(emptyReader, runtime.New())
		engine.logger = logger
		err := engine.AddScript(newScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on(
					'http', 
					function(p) {
						console.log(p.foo);
					}
				);
			}
		`))
		r.NoError(t, err)

		engine.Run("http", p)

		r.Equal(t, "bar", msg)
	})
}

func TestJsOpen(t *testing.T) {
	t.Parallel()
	t.Run("fileExists", func(t *testing.T) {
		t.Parallel()
		var msg string
		logger := &testLogger{
			info: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		reader := &testReader{readFunc: func(cfg *common.Config) error {
			cfg.Raw = []byte("foobar")
			return nil
		}}

		engine := New(reader, runtime.New())
		engine.logger = logger
		err := engine.AddScript(newScript("./test.js", `
			let file = open('test.txt');
			console.log(file);
			export default function() {}
		`))
		r.NoError(t, err)
		r.Equal(t, "foobar", msg)
	})
	t.Run("fileNotExists", func(t *testing.T) {
		t.Parallel()

		reader := &testReader{readFunc: func(cfg *common.Config) error {
			return errors.New("file not found")
		}}

		engine := New(reader, runtime.New())
		err := engine.AddScript(newScript("./test.js", `
			let file = open('test.txt');
			export default function() {}
		`))
		r.True(t, strings.HasPrefix(err.Error(), "GoError: file not found"), "file not found")
	})
	t.Run("require nested with update", func(t *testing.T) {
		t.Parallel()
		foo := `const {bar} = require('bar'); export let foo = bar`
		bar := `export let bar = 'bar'; export let xy = 'xy'`
		var barFile *common.Config

		reader := &testReader{readFunc: func(cfg *common.Config) error {
			switch s := cfg.Url.String(); {
			case strings.HasSuffix(s, "foo.js"):
				cfg.Raw = []byte(foo)
				return nil
			case strings.HasSuffix(s, "bar.js"):
				barFile = cfg
				cfg.Raw = []byte(bar)
				return nil
			}
			return errors.New("file not found")
		}}

		engine := New(reader, runtime.New())
		s := newScript("./test.js", `
			import {foo} from 'foo'
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {name: foo}});
			}
		`)
		s.AddListener("", func(config *common.Config) {
			err := engine.AddScript(config)
			r.NoError(t, err)
		})
		err := engine.AddScript(s)
		r.NoError(t, err)

		summaries := engine.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		r.Equal(t, "bar", summary.Tags["name"])

		bar = `export let bar = 'foobar'`
		barFile.Checksum = []byte("foobar")
		barFile.Changed()

		summaries = engine.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary = summaries[0]
		r.Equal(t, "foobar", summary.Tags["name"])
	})
}

func newScript(path, src string) *common.Config {
	s := common.NewConfig(mustParse(path), common.WithData(&script.Script{Code: src, Filename: path}))
	s.Raw = []byte(src)
	return s
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
