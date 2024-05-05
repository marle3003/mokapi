package engine_test

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/dynamic/script"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"mokapi/js/require"
	"net/url"
	"strings"
	"testing"
)

func TestJsScriptEngine(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", "export default function(){}"))
		r.NoError(t, err)
		r.Equal(t, 0, e.Scripts(), "no events and jobs, script should be closed")
	})
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", ""))
		r.NoError(t, err)
		r.Equal(t, 0, e.Scripts(), "no events and jobs, script should be closed")
	})
	t.Run("typescript", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.ts", "const msg: string = 'Hello World';"))
		r.NoError(t, err)
		r.Equal(t, 0, e.Scripts(), "no events and jobs, script should be closed")
	})
	t.Run("typescript async default function", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", "export default async function(){ setTimeout(() => { mokapi.every('1m', function() {}) }, 500)}"))
		r.NoError(t, err)
		r.Equal(t, 1, e.Scripts(), "no events and jobs, script should be closed")
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
				registered := false
				scheduler := &enginetest.Scheduler{
					EveryFunc: func(every string, handler func(), opt common.JobOptions) (engine.Job, error) {
						registered = true
						r.Equal(t, "1m", every)
						return nil, nil
					},
				}
				e := enginetest.NewEngine(engine.WithScheduler(scheduler))
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1m', function() {});
					}
				`))
				r.NoError(t, err)
				r.Equal(t, 1, e.Scripts(), "script length not 1")

				r.True(t, registered)
			},
		},
		{
			name: "job runs immediately",
			test: func(t *testing.T) {
				registered := false
				scheduler := &enginetest.Scheduler{
					EveryFunc: func(every string, handler func(), opt common.JobOptions) (engine.Job, error) {
						registered = true
						r.Equal(t, false, opt.SkipImmediateFirstRun)
						return nil, nil
					},
				}
				e := enginetest.NewEngine(engine.WithScheduler(scheduler))
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1h', function() {});
					}
				`))
				r.NoError(t, err)
				r.True(t, registered)
			},
		},
		{
			name: "job runs only 2 times",
			test: func(t *testing.T) {
				registered := false
				scheduler := &enginetest.Scheduler{
					EveryFunc: func(every string, handler func(), opt common.JobOptions) (engine.Job, error) {
						registered = true
						r.Equal(t, 2, opt.Times)
						return nil, nil
					},
				}
				e := enginetest.NewEngine(engine.WithScheduler(scheduler))
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('100ms', function() {}, { times: 2 });
					}
				`))
				r.NoError(t, err)
				r.True(t, registered)
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
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {}
		`))
		r.NoError(t, err)
		r.Equal(t, 0, e.Scripts(), "script length not 1")
	})
	t.Run("withoutSummary", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					return false
				});
			}
		`))
		r.NoError(t, err)
		r.Equal(t, 1, e.Scripts(), "script length not 1")

		summaries := e.Run("http")

		r.Len(t, summaries, 0, "summary length not 0")
	})
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function(request, response) {
					return true
				});
			}
		`))
		r.NoError(t, err)
		r.Equal(t, 1, e.Scripts(), "script length not 1")

		summaries := e.Run("http", &struct{}{}, &struct{}{})

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		// tags
		r.Equal(t, "test.js", summary.Tags["name"], "tag name not correct")
		r.Equal(t, "http", summary.Tags["event"], "tag event not correct")
	})
	t.Run("duration", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					sleep(1000);
					return true
				});
			}
		`))
		r.NoError(t, err)

		summaries := e.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		r.GreaterOrEqual(t, summary.Duration, int64(1000), "sleep")
	})
	t.Run("duration as string", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on, sleep} from 'mokapi'
			export default function() {
				on('http', function() {
					sleep('1s');
					return true
				});
			}
		`))
		r.NoError(t, err)

		summaries := e.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		r.GreaterOrEqual(t, summary.Duration, int64(1000), "sleep")
	})
	t.Run("tag name", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {'name': 'foobar'}});
			}
		`))
		r.NoError(t, err)

		summaries := e.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		r.Equal(t, "foobar", summaries[0].Tags["name"], "tag name not correct")
	})
	t.Run("custom tag", func(t *testing.T) {
		t.Parallel()
		e := enginetest.NewEngine()
		err := e.AddScript(newScript("test.js", `
			import {on} from 'mokapi'
			export default function() {
				on('http', function() {return true}, {tags: {'foo': 'bar'}});
			}
		`))
		r.NoError(t, err)

		summaries := e.Run("http")

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
		logger := &enginetest.Logger{
			InfoFunc: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		e := enginetest.NewEngine(engine.WithLogger(logger))
		err := e.AddScript(newScript("test.js", `
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

		e.Run("http", p)

		r.Equal(t, "bar", msg)
	})
}

func TestJsOpen(t *testing.T) {
	t.Parallel()
	t.Run("fileExists", func(t *testing.T) {
		t.Parallel()
		var msg string
		logger := &enginetest.Logger{
			InfoFunc: func(args ...interface{}) {
				msg = fmt.Sprintf("%v", args[0])
			},
		}

		reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
			return &dynamic.Config{Raw: []byte("foobar")}, nil
		})

		e := enginetest.NewEngine(engine.WithLogger(logger), engine.WithReader(reader))
		err := e.AddScript(newScript("./test.js", `
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

		e := enginetest.NewEngine(engine.WithReader(reader))
		err := e.AddScript(newScript("./test.js", `
			let file = open('test.txt');
			export default function() {}
		`))
		r.True(t, strings.HasPrefix(err.Error(), "GoError: file not found"), "file not found")
	})
	t.Run("require nested with change file", func(t *testing.T) {
		t.Parallel()
		foo := `const { bar } = require('bar'); export let foo = bar`
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

		e := enginetest.NewEngine(engine.WithReader(reader), engine.WithScriptLoader(
			engine.ScriptLoaderFunc(func(file *dynamic.Config, host common.Host) (common.Script, error) {
				registry, err := require.NewRegistry(host.OpenFile)
				r.NoError(t, err)
				js.RegisterNativeModules(registry)

				return jstest.New(js.WithFile(file), js.WithHost(host), js.WithRegistry(registry))
			}),
		))
		s := newScript("./test.js", `
			import { foo } from 'foo'
			import { on } from 'mokapi'
			export default function() {
				on('http', function() { return true }, { tags: { name: foo } });
			}
		`)
		s.Listeners.Add("", func(config *dynamic.Config) {
			err := e.AddScript(config)
			r.NoError(t, err)
		})
		err := e.AddScript(s)
		r.NoError(t, err)

		summaries := e.Run("http")

		r.Len(t, summaries, 1, "summary length not 1")
		summary := summaries[0]
		r.Equal(t, "bar", summary.Tags["name"])

		bar = `export let bar = 'foobar'`
		barFile.Info.Checksum = []byte("foobar")
		barFile.Listeners.Invoke(barFile)

		summaries = e.Run("http")

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

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
