package engine

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/runtime"
	"net/url"
	"strings"
	"testing"
	"time"
)

var reader = dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
	return nil, nil
})

func TestJsScriptEngine(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
		err := engine.AddScript(newScript("test.js", "export default function(){}"))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 0, "no events and jobs, script should be closed")
	})
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
		err := engine.AddScript(newScript("test.js", ""))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 0, "no events and jobs, script should be closed")
	})
}

func TestJsEvery(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "job is registered",
			test: func(t *testing.T) {
				engine := New(reader, runtime.New(), static.JsConfig{}, false)
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
			},
		},
		{
			name: "job runs immediately",
			test: func(t *testing.T) {
				engine := New(reader, runtime.New(), static.JsConfig{}, false)
				go engine.Start()
				defer engine.Close()
				err := engine.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1h', function() {});
					}
				`))
				r.NoError(t, err)
				r.Len(t, engine.scripts, 1, "script length not 1")

				time.Sleep(500 * time.Millisecond)

				r.Equal(t, 1, engine.scripts["test.js"].jobs[0].RunCount(), "job run count")
			},
		},
		{
			name: "job runs only 2 times",
			test: func(t *testing.T) {
				engine := New(reader, runtime.New(), static.JsConfig{}, false)
				go engine.Start()
				defer engine.Close()
				err := engine.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('100ms', function() {}, { times: 2 });
					}
				`))
				r.NoError(t, err)
				time.Sleep(500 * time.Millisecond)

				r.Equal(t, 2, engine.scripts["test.js"].jobs[0].RunCount(), "job run count")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestJsOn(t *testing.T) {
	t.Parallel()
	t.Run("noEvent", func(t *testing.T) {
		t.Parallel()
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
		err := engine.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {}
		`))
		r.NoError(t, err)
		r.Len(t, engine.scripts, 0, "script length not 1")
	})
	t.Run("withoutSummary", func(t *testing.T) {
		t.Parallel()
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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
	t.Run("duration as string", func(t *testing.T) {
		t.Parallel()
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
		err := engine.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					sleep('1s');
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
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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
		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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

		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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

		reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
			return &dynamic.Config{Raw: []byte("foobar")}, nil
		})

		engine := New(reader, runtime.New(), static.JsConfig{}, false)
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

		reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
			return nil, fmt.Errorf("file not found")
		})

		engine := New(reader, runtime.New(), static.JsConfig{}, false)
		err := engine.AddScript(newScript("./test.js", `
			let file = open('test.txt');
			export default function() {}
		`))
		r.True(t, strings.HasPrefix(err.Error(), "GoError: file not found"), "file not found")
	})
	t.Run("require nested with change file", func(t *testing.T) {
		t.Parallel()
		foo := `const {bar} = require('bar'); export let foo = bar`
		bar := `export let bar = 'bar'; export let xy = 'xy'`
		var barFile *dynamic.Config

		reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
			cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
			switch s := cfg.Info.Url.String(); {
			case strings.HasSuffix(s, "foo.js"):
				cfg.Raw = []byte(foo)
				return cfg, nil
			case strings.HasSuffix(s, "bar.js"):
				barFile = cfg
				cfg.Raw = []byte(bar)
				return cfg, nil
			}
			return nil, fmt.Errorf("file not found")
		})

		engine := New(reader, runtime.New(), static.JsConfig{}, false)
		s := newScript("./test.js", `
			import {foo} from 'foo'
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {name: foo}});
			}
		`)
		s.Listeners.Add("", func(config *dynamic.Config) {
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
		barFile.Info.Checksum = []byte("foobar")
		barFile.Listeners.Invoke(barFile)

		summaries = engine.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary = summaries[0]
		r.Equal(t, "foobar", summary.Tags["name"])
	})
}

func newScript(path, src string) *dynamic.Config {
	return &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: mustParse(path)},
		Raw:  []byte(src),
		Data: &script.Script{Code: src, Filename: path},
	}
}

type testLogger struct {
	info  func(args ...interface{})
	warn  func(args ...interface{})
	error func(args ...interface{})
	debug func(args ...interface{})
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

func (tl *testLogger) Debug(args ...interface{}) {
	tl.debug(args...)
}
