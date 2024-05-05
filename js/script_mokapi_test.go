package js_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestScript_Mokapi_Date(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "now default",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { date } from 'mokapi'
						 export default function() {
						  	return date({timestamp:  new Date(Date.UTC(2022, 5, 9, 12, 0, 0, 0)).getTime()}); // january is 0
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				expected := time.Date(2022, 6, 9, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
				r.Equal(t, expected, i.String())
			},
		},
		{
			name: "utc time ends with Z",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { date } from 'mokapi'
						 export default function() {
						  	return date()
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
			},
		},
		{
			name: "utc time ends with Z",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { date } from 'mokapi'
						 export default function() {
						  	return date({timestamp: Date.now()})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
			},
		},
		{
			name: "custom format",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { date } from 'mokapi'
						 export default function() {
						  	return date({layout: 'DateTime', timestamp: new Date(Date.UTC(2022, 5, 9, 12, 0, 0, 0)).getTime()})
						 }`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_Every(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "timer",
			test: func(t *testing.T, host *enginetest.Host) {
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "1s", every)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "but one time",
			test: func(t *testing.T, host *enginetest.Host) {
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, 1, opt.Times)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, {times: 1})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "skip immediate first run ",
			test: func(t *testing.T, host *enginetest.Host) {
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					r.True(t, opt.SkipImmediateFirstRun)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, { skipImmediateFirstRun : true })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "with custom tags",
			test: func(t *testing.T, host *enginetest.Host) {
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "bar", opt.Tags["foo"])
				}
				s, err := jstest.New(jstest.WithSource(
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, {tags: {foo: 'bar'}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "tags set to null",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { every } from 'mokapi'
						 export default function() {
						  	every('1s', function() {}, {tags: null})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "run function",
			test: func(t *testing.T, host *enginetest.Host) {
				host.EveryFunc = func(every string, do func(), opt common.JobOptions) {
					do()
				}
				s, err := jstest.New(jstest.WithSource(
					`import { every } from 'mokapi'
						 export default function() {
							let counter = 1
						  	every('1s', function() {counter++}, {tags: {foo: 'bar'}})
							return counter
						 }`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_Cron(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "timer",
			test: func(t *testing.T, host *enginetest.Host) {
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "0/1 0 0 ? * * *", every)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "but one time",
			test: func(t *testing.T, host *enginetest.Host) {
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, 1, opt.Times)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, {times: 1})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "skip immediate first run",
			test: func(t *testing.T, host *enginetest.Host) {
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					r.True(t, opt.SkipImmediateFirstRun)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, { skipImmediateFirstRun: true })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "with custom tags",
			test: func(t *testing.T, host *enginetest.Host) {
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					r.Equal(t, "bar", opt.Tags["foo"])
				}
				s, err := jstest.New(jstest.WithSource(
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, {tags: {foo: 'bar'}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "tags set to null",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { cron } from 'mokapi'
						 export default function() {
						  	cron('0/1 0 0 ? * * *', function() {}, {tags: null})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "run function",
			test: func(t *testing.T, host *enginetest.Host) {
				host.CronFunc = func(every string, do func(), opt common.JobOptions) {
					do()
				}
				s, err := jstest.New(jstest.WithSource(
					`import { cron } from 'mokapi'
						 export default function() {
							let counter = 1
						  	cron('0/1 0 0 ? * * *', function() {counter++}, {tags: {foo: 'bar'}})
							return counter
						 }`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_On_Http(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "event",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "tags",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "bar", tags["foo"])
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {}, {tags: {foo: 'bar'}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "tags set to null",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {}, {tags: null})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
			},
		},
		{
			name: "run function",
			test: func(t *testing.T, host *enginetest.Host) {
				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						let counter = 0
						 export default function() {
						  	on('http', function(arg) {counter = arg})
						 }`),
					js.WithHost(host))

				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				doFunc(10)

				var v goja.Value
				err = s.RunFunc(func(vm *goja.Runtime) {
					v = vm.Get("counter")
				})

				time.Sleep(100 * time.Millisecond)
				r.Equal(t, int64(10), v.ToInteger())
				s.Close()
			},
		},
		{
			name: "return value default is false",
			test: func(t *testing.T, host *enginetest.Host) {
				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := doFunc()
				r.NoError(t, err)
				r.False(t, b, "default return value should be false")
				s.Close()
			},
		},
		{
			name: "return value true",
			test: func(t *testing.T, host *enginetest.Host) {
				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {return true})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				b, err := doFunc()
				r.NoError(t, err)
				r.True(t, b, "return value should be true")
				r.NoError(t, err)
				s.Close()
			},
		},
		{
			name: "on error",
			test: func(t *testing.T, host *enginetest.Host) {
				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function() {throw new Error('test error')})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := doFunc()
				r.EqualError(t, err, "Error: test error at test.js:3:46(3)")
				r.False(t, b, "return value should be false on error")
				s.Close()
			},
		},
		{
			name: "access struct by dot notation",
			test: func(t *testing.T, host *enginetest.Host) {
				data := &struct {
					ShipDate string
				}{
					ShipDate: "2022-01-01",
				}

				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function(data) {
								return data.shipDate === '2022-01-01'
							})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := doFunc(data)
				r.NoError(t, err)
				r.True(t, b, "return value should be true")
				s.Close()
			},
		},
		{
			name: "access kebab case property by bracket notation",
			test: func(t *testing.T, host *enginetest.Host) {
				data := &struct {
					Ship_date string `json:"ship-date"` // can be accessed via obj['ship-date'] in javascript
				}{
					Ship_date: "2022-01-01",
				}

				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function(data) {
								return data['ship-date'] === '2022-01-01'
							})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := doFunc(data)
				r.NoError(t, err)
				r.True(t, b, "return value should be true")
				s.Close()
			},
		},
		{
			name: "access map by object by dot notation",
			test: func(t *testing.T, host *enginetest.Host) {
				data := map[string]string{"foo": "bar"}

				var doFunc func(args ...interface{}) (bool, error)
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					doFunc = do
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('http', function(data) {
								return data.foo === 'bar'
							})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := doFunc(data)
				r.NoError(t, err)
				r.True(t, b, "return value should be true")
				s.Close()
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_On_Kafka(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "event",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "kafka", event)
				}
				s, err := jstest.New(jstest.WithSource(
					`import { on } from 'mokapi'
						 export default function() {
						  	on('kafka', function(message) {
								
							})
						 }`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_Env(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "env",
			test: func(t *testing.T, host *enginetest.Host) {
				os.Setenv("foo", "bar")
				s, err := jstest.New(jstest.WithSource(
					`import { env } from 'mokapi'
						 export default function() {
						  	return env('foo')
						 }`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_Sleep(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "sleep",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import { sleep } from 'mokapi'
						 export default function() {
							sleep(300);
						}`),
					js.WithHost(host))
				r.NoError(t, err)
				start := time.Now()
				_, err = s.RunDefault()
				r.NoError(t, err)
				duration := time.Now().Sub(start).Milliseconds()
				r.Greater(t, duration, int64(300))
			},
		},
		{
			name: "sleep with string",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import { sleep } from 'mokapi'
						 export default function() {
							sleep('300ms');
						}`),
					js.WithHost(host))
				r.NoError(t, err)
				start := time.Now()
				_, err = s.RunDefault()
				r.NoError(t, err)
				duration := time.Now().Sub(start).Milliseconds()
				r.Greater(t, duration, int64(300))
			},
		},
		{
			name: "sleep invalid time format",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import { sleep } from 'mokapi'
						 export default function() {
							sleep('300-');
						}`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.EqualError(t, err, "time: unknown unit \"-\" in duration \"300-\" at mokapi/js/mokapi.(*Module).Sleep-fm (native)")
			},
		},
		{
			name: "catch exception sleep invalid time format",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import { sleep } from 'mokapi'
						 export default function() {
							try {
							sleep('300-');
							} catch(e){
								return e
							}
						}`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}

func TestScript_Mokapi_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "default encoding",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { marshal } from 'mokapi'
						 export default function() {
						  	return marshal({ username: 'foo' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, `{"username":"foo"}`, i.String())
			},
		},
		{
			name: "with schema",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { marshal } from 'mokapi'
						 export default function() {
						  	return marshal({ username: 'foo' }, { 
								schema: { 
									type: 'object',
									properties: {
										username: {
											type: 'string'
										}
									}
								}
							})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, `{"username":"foo"}`, i.String())
			},
		},
		{
			name: "with content type xml",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { marshal } from 'mokapi'
						 export default function() {
						  	return marshal({ username: 'foo' }, { 
								schema: { 
									type: 'object',
									xml: { name: 'user' }
								},
								contentType: 'application/xml'
							})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, `<user><username>foo</username></user>`, i.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t, &enginetest.Host{})
		})
	}
}
