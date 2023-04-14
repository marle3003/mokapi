package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
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
					host)
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
					host)
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
					host)
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
					host)
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
