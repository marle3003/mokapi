package js

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestScript(t *testing.T) {
	t.Parallel()
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		s, err := New(newScript("", ""), &testHost{}, static.JsConfig{})
		r.NoError(t, err)
		err = s.Run()
		r.NoError(t, err)
	})
	t.Run("null", func(t *testing.T) {
		t.Parallel()
		s, err := New(newScript("", "exports = null"), &testHost{}, static.JsConfig{})
		r.NoError(t, err)
		err = s.Run()
		r.NoError(t, err)
	})
	t.Run("emptyFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New(newScript("test", `export default function() {}`), &testHost{}, static.JsConfig{})
		r.NoError(t, err)
		r.NoError(t, s.Run())
	})
	t.Run("console.log", func(t *testing.T) {
		t.Parallel()
		host := &testHost{}
		host.info = func(args ...interface{}) {
			r.Equal(t, "foo", args[0])
		}
		s, err := New(newScript("test", `export default function() {console.log("foo")}`), host, static.JsConfig{})
		r.NoError(t, err)
		r.NoError(t, s.Run())
	})
	t.Run("console.warn", func(t *testing.T) {
		t.Parallel()
		host := &testHost{}
		host.warn = func(args ...interface{}) {
			r.Equal(t, "foo", args[0])
		}
		s, err := New(newScript("test", `export default function() {console.warn("foo")}`), host, static.JsConfig{})
		r.NoError(t, err)
		r.NoError(t, s.Run())
	})
	t.Run("console.err", func(t *testing.T) {
		t.Parallel()
		host := &testHost{}
		host.error = func(args ...interface{}) {
			r.Equal(t, "foo", args[0])
		}
		s, err := New(newScript("test", `export default function() {console.error("foo")}`), host, static.JsConfig{})
		r.NoError(t, err)
		r.NoError(t, s.Run())
	})
	t.Run("returnValueFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New(newScript("test", `export default function() {return 2}`), &testHost{}, static.JsConfig{})
		r.NoError(t, err)
		r.NoError(t, s.Run())
		err = s.Run()
		f, ok := goja.AssertFunction(s.exports.ToObject(s.runtime).Get("default"))
		r.True(t, ok)
		v, err := f(goja.Undefined())
		r.NoError(t, err)
		r.Equal(t, int64(2), v.ToInteger())
	})
	t.Run("customFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New(newScript("test", `function custom() {return 2}; export {custom}`), &testHost{}, static.JsConfig{})
		r.NoError(t, err)
		r.NoError(t, s.Run())
		f, ok := goja.AssertFunction(s.exports.ToObject(s.runtime).Get("custom"))
		r.True(t, ok)
		v, err := f(goja.Undefined())
		r.NoError(t, err)
		r.Equal(t, int64(2), v.ToInteger())
	})
	t.Run("interrupt", func(t *testing.T) {
		t.Parallel()
		s, err := New(newScript("test", `export default function() {while(true) {}}`), &testHost{}, static.JsConfig{})
		r.NoError(t, err)
		ch := make(chan bool)
		go func() {
			ch <- true
			err := s.Run()
			iErr := err.(*goja.InterruptedError)
			r.True(t, strings.HasPrefix(iErr.String(), "closing"), fmt.Sprintf("error prefix expected closing but got: %v", iErr.String()))
		}()

		<-ch
		<-time.NewTimer(time.Duration(1) * time.Second).C
		s.Close()
	})
	t.Run("warn deprecated module", func(t *testing.T) {
		t.Parallel()
		host := &testHost{}
		s, err := New(newScript("test", `import http from 'http'
											export default function() {}`), host, static.JsConfig{})
		r.NoError(t, err)
		var warn interface{}
		host.warn = func(args ...interface{}) {
			warn = args[0]
		}
		err = s.Run()
		r.NoError(t, err)
		r.Equal(t, "deprecated module http: Please use mokapi/http instead: test", warn)
		s.Close()
	})
}

type testHost struct {
	openFile    func(file, hint string) (string, string, error)
	open        func(file, hint string) (*dynamic.Config, error)
	info        func(args ...interface{})
	warn        func(args ...interface{})
	error       func(args ...interface{})
	debug       func(args ...interface{})
	httpClient  *testClient
	kafkaClient *kafkaClient
	every       func(every string, do func(), opt common.JobOptions)
	cron        func(every string, do func(), opt common.JobOptions)
	on          func(event string, do func(args ...interface{}) (bool, error), tags map[string]string)
}

func (th *testHost) Info(args ...interface{}) {
	if th.info != nil {
		th.info(args...)
	}
}

func (th *testHost) Warn(args ...interface{}) {
	if th.warn != nil {
		th.warn(args...)
	}
}

func (th *testHost) Error(args ...interface{}) {
	if th.error != nil {
		th.error(args...)
	}
}

func (th *testHost) Debug(args ...interface{}) {
	if th.debug != nil {
		th.debug(args...)
	}
}

func (th *testHost) OpenFile(file, hint string) (*dynamic.Config, error) {
	if th.openFile != nil {
		p, src, err := th.openFile(file, hint)
		if err != nil {
			return nil, err
		}
		return &dynamic.Config{Raw: []byte(src), Info: dynamic.ConfigInfo{Url: mustParse(p)}}, nil
	}
	if th.open != nil {
		return th.open(file, hint)
	}
	return nil, fmt.Errorf("file %v not found (hint: %v)", file, hint)
}

func (th *testHost) Every(every string, do func(), opt common.JobOptions) (int, error) {
	if th.every != nil {
		th.every(every, do, opt)
	}
	return 0, nil
}

func (th *testHost) Cron(expr string, do func(), opt common.JobOptions) (int, error) {
	if th.cron != nil {
		th.cron(expr, do, opt)
	}
	return 0, nil
}

func (th *testHost) On(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
	if th.on != nil {
		th.on(event, do, tags)
	}
}

func (th *testHost) HttpClient() common.HttpClient {
	return th.httpClient
}

func (th *testHost) KafkaClient() common.KafkaClient {
	return th.kafkaClient
}

type testClient struct {
	req    *http.Request
	doFunc func(request *http.Request) (*http.Response, error)
}

func (c *testClient) Do(request *http.Request) (*http.Response, error) {
	c.req = request
	if c.doFunc != nil {
		return c.doFunc(request)
	}
	return &http.Response{}, nil
}

type kafkaClient struct {
	produce func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error)
}

func (c *kafkaClient) Produce(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
	if c.produce != nil {
		return c.produce(args)
	}
	return nil, nil
}

func (th *testHost) Cancel(jobId int) error {
	return nil
}

func (th *testHost) Name() string {
	return "test host"
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
