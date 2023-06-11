package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/engine/common"
	"os"
	"strings"
	"testing"
	"time"
)

func TestScript_Mokapi_Date(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"now default",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { date } from 'mokapi'
						 export default function() {
						  	return date({timestamp:  new Date(Date.UTC(2022, 5, 9, 12, 0, 0, 0)).getTime()}); // january is 0
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				expected := time.Date(2022, 6, 9, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
				r.Equal(t, expected, i.String())
			},
		},
		{
			"utc time ends with Z",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { date } from 'mokapi'
						 export default function() {
						  	return date()
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
			},
		},
		{
			"utc time ends with Z",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { date } from 'mokapi'
						 export default function() {
						  	return date({timestamp: Date.now()})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
			},
		},
		{
			"custom format",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { date } from 'mokapi'
						 export default function() {
						  	return date({layout: 'DateTime', timestamp: new Date(Date.UTC(2022, 5, 9, 12, 0, 0, 0)).getTime()})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "2022-06-09 12:00:00", i.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}

func TestScript_Mokapi_Every(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"timer",
			func(t *testing.T, host *testHost) {
				host.every = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "1s", every)
				}
				s, err := New("",
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"but one time",
			func(t *testing.T, host *testHost) {
				host.every = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, 1, opt.Times)
				}
				s, err := New("",
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, {times: 1})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"with custom tags",
			func(t *testing.T, host *testHost) {
				host.every = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "bar", opt.Tags["foo"])
				}
				s, err := New("",
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, {tags: {foo: 'bar'}})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"tags set to null",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, {tags: null})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"run function",
			func(t *testing.T, host *testHost) {
				host.every = func(every string, do func(), opt common.JobOptions) {
					do()
				}
				s, err := New("",
					`import { every } from 'mokapi'
						 export default function() {
							let counter = 1
						  	every('1s', function() {counter++}, {tags: {foo: 'bar'}})
							return counter
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, int64(2), v.ToInteger())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}

func TestScript_Mokapi_Cron(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"timer",
			func(t *testing.T, host *testHost) {
				host.cron = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "0/1 0 0 ? * * *", every)
				}
				s, err := New("",
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"but one time",
			func(t *testing.T, host *testHost) {
				host.cron = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, 1, opt.Times)
				}
				s, err := New("",
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, {times: 1})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"with custom tags",
			func(t *testing.T, host *testHost) {
				host.cron = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "bar", opt.Tags["foo"])
				}
				s, err := New("",
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, {tags: {foo: 'bar'}})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"tags set to null",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, {tags: null})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"run function",
			func(t *testing.T, host *testHost) {
				host.cron = func(every string, do func(), opt common.JobOptions) {
					do()
				}
				s, err := New("",
					`import { cron } from 'mokapi'
						 export default function() {
							let counter = 1
						  	cron('0/1 0 0 ? * * *', function() {counter++}, {tags: {foo: 'bar'}})
							return counter
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, int64(2), v.ToInteger())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}

func TestScript_Mokapi_On(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"event",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
				}
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"tags",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "bar", tags["foo"])
				}
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {}, {tags: {foo: 'bar'}})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"tags set to null",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {}, {tags: null})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"run function",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					do(10)
				}
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
							let counter = 0
						  	on('http', function(arg) {counter = arg})
							return counter
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, int64(10), v.ToInteger())
			},
		},
		{
			"return value default is false",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					b, err := do()
					r.NoError(t, err)
					r.False(t, b, "default return value should be false")
				}
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"return value true",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					b, err := do()
					r.NoError(t, err)
					r.True(t, b, "return value should be true")
				}
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {return true})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			"on error",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					b, err := do()
					r.EqualError(t, err, "Error: test error at <eval>:3:46(3)")
					r.False(t, b, "return value should be false on error")
				}
				s, err := New("",
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {throw new Error('test error')})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}

func TestScript_Mokapi_Env(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"env",
			func(t *testing.T, host *testHost) {
				os.Setenv("foo", "bar")
				s, err := New("",
					`import { env } from 'mokapi'
						 export default function() {
						  	return env('foo')
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}

func TestScript_Mokapi_Open(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"open",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := New("",
					`import { open } from 'mokapi'
						 export default function() {
						  	return open('foo')
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.String())
			},
		},
		{
			"file not found",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "", fmt.Errorf("test error")
				}
				s, err := New("",
					`import { open } from 'mokapi'
						 export default function() {
						  	return open('foo')
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}

func TestScript_Mokapi_Sleep(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"sleep",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := New("",
					`import { sleep } from 'mokapi'
						 export default function() {
							sleep(300);
						}`,
					host, static.JsConfig{})
				r.NoError(t, err)
				start := time.Now()
				_, err = s.RunDefault()
				r.NoError(t, err)
				duration := time.Now().Sub(start).Milliseconds()
				r.Greater(t, duration, int64(300))
			},
		},
		{
			"sleep with string",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := New("",
					`import { sleep } from 'mokapi'
						 export default function() {
							sleep('300ms');
						}`,
					host, static.JsConfig{})
				r.NoError(t, err)
				start := time.Now()
				_, err = s.RunDefault()
				r.NoError(t, err)
				duration := time.Now().Sub(start).Milliseconds()
				r.Greater(t, duration, int64(300))
			},
		},
		{
			"sleep invalid time format",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := New("",
					`import { sleep } from 'mokapi'
						 export default function() {
							sleep('300-');
						}`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.EqualError(t, err, "time: unknown unit \"-\" in duration \"300-\" at reflect.methodValueCall (native)")
			},
		},
		{
			"catch exception sleep invalid time format",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := New("",
					`import { sleep } from 'mokapi'
						 export default function() {
							try {
							sleep('300-');
							} catch(e){
								return e
							}
						}`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "time: unknown unit \"-\" in duration \"300-\"", v.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}
