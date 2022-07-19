package js

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestScript(t *testing.T) {
	host := &testHost{}

	t.Parallel()
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		s, err := New("", "", host)
		r.NoError(t, err)
		err = s.Run()
		r.NoError(t, err)
	})
	t.Run("null", func(t *testing.T) {
		t.Parallel()
		s, err := New("", "exports = null", host)
		r.NoError(t, err)
		err = s.Run()
		r.NoError(t, err)
	})
	t.Run("emptyFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {}`, host)
		r.NoError(t, err)
		r.NoError(t, s.Run())
	})
	t.Run("console.log", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {console.log("foo")}`, host)
		r.NoError(t, err)
		r.NoError(t, s.Run())
	})
	t.Run("returnValueFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {return 2}`, host)
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
		s, err := New("test", `function custom() {return 2}; export {custom}`, host)
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
		s, err := New("test", `export default function() {while(true) {}}`, host)
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
}

func TestScript_Generator(t *testing.T) {
	host := &testHost{}

	t.Parallel()
	t.Run("nil", func(t *testing.T) {
		t.Parallel()

		s, err := New("",
			`
import {fake} from 'faker'
export default function() {
  var s = fake({type: 'string'})
return s
}`,
			host)
		r.NoError(t, err)
		err = s.Run()
		r.NoError(t, err)
	})
}

type testHost struct {
	common.Host
	openFile    func(file, hint string) (string, string, error)
	openScript  func(file, hint string) (string, string, error)
	info        func(args ...interface{})
	httpClient  *testClient
	kafkaClient *kafkaClient
}

func (th *testHost) Info(args ...interface{}) {
	if th.info != nil {
		th.info(args...)
	}
}

func (th *testHost) OpenFile(file, hint string) (string, string, error) {
	if th.openFile != nil {
		return th.openFile(file, hint)
	}
	return "", "", nil
}

func (th *testHost) OpenScript(file, hint string) (string, string, error) {
	if th.openScript != nil {
		return th.openScript(file, hint)
	}
	return "", "", nil
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
	produce func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error)
}

func (c *kafkaClient) Produce(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
	if c.produce != nil {
		return c.produce(cluster, topic, partition, key, value, headers)
	}
	return nil, nil, nil
}
