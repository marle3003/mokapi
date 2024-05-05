package js_test

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestScript(t *testing.T) {
	t.Parallel()
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		s, err := jstest.New(jstest.WithSource(""), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		_, err = s.RunDefault()
		r.Equal(t, err, js.NoDefaultFunction)
		s.Close()
	})
	t.Run("null", func(t *testing.T) {
		t.Parallel()
		s, err := jstest.New(jstest.WithSource("exports = null"), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		err = s.Run()
		r.NoError(t, err)
		s.Close()
	})
	t.Run("emptyFunction", func(t *testing.T) {
		t.Parallel()
		s, err := jstest.New(jstest.WithSource(`export default function() {}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		_, err = s.RunDefault()
		r.NoError(t, err)
		s.Close()
	})
	t.Run("console.log", func(t *testing.T) {
		t.Parallel()
		host := &enginetest.Host{}
		host.InfoFunc = func(args ...interface{}) {
			r.Equal(t, "foo", args[0])
		}
		s, err := jstest.New(jstest.WithSource(`export default function() {console.log("foo")}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		_, err = s.RunDefault()
		r.NoError(t, err)
		s.Close()
	})
	t.Run("console.warn", func(t *testing.T) {
		t.Parallel()
		host := &enginetest.Host{}
		host.WarnFunc = func(args ...interface{}) {
			r.Equal(t, "foo", args[0])
		}
		s, err := jstest.New(jstest.WithSource(`export default function() {console.warn("foo")}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		_, err = s.RunDefault()
		r.NoError(t, err)
		s.Close()
	})
	t.Run("console.err", func(t *testing.T) {
		t.Parallel()
		host := &enginetest.Host{}
		host.ErrorFunc = func(args ...interface{}) {
			r.Equal(t, "foo", args[0])
		}
		s, err := jstest.New(jstest.WithSource(`export default function() {console.error("foo")}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		_, err = s.RunDefault()
		r.NoError(t, err)
		s.Close()
	})
	t.Run("returnValueFunction", func(t *testing.T) {
		t.Parallel()
		s, err := jstest.New(jstest.WithSource(`export default function() {return 2}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		returnValue, err := s.RunDefault()
		r.NoError(t, err)
		r.Equal(t, int64(2), returnValue.ToInteger())
		s.Close()
	})
	t.Run("customFunction", func(t *testing.T) {
		t.Parallel()
		s, err := jstest.New(jstest.WithSource(`function custom() {return 2}; export {custom}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		err = s.RunFunc(func(vm *goja.Runtime) {
			exports := vm.Get("exports").ToObject(vm)
			f, _ := goja.AssertFunction(exports.Get("custom"))
			v, err := f(goja.Undefined())
			r.NoError(t, err)
			r.Equal(t, int64(2), v.ToInteger())
		})
		r.NoError(t, err)
		s.Close()
	})
	t.Run("interrupt", func(t *testing.T) {
		t.Parallel()
		s, err := jstest.New(jstest.WithSource(`export default function() {while(true) {}}`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		ch := make(chan bool)
		go func() {
			ch <- true
			_, err := s.RunDefault()
			var iErr *goja.InterruptedError
			errors.As(err, &iErr)
			r.True(t, strings.HasPrefix(iErr.String(), "closing"), fmt.Sprintf("error prefix expected closing but got: %v", iErr.String()))
		}()

		<-ch
		<-time.NewTimer(time.Duration(1) * time.Second).C
		s.Close()
	})
	t.Run("access process environment variable", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOKAPI_IS_AWESOME", "true")
		defer os.Unsetenv("MOKAPI_IS_AWESOME")

		s, err := jstest.New(jstest.WithSource(`export default function() { return process.env['MOKAPI_IS_AWESOME'] }`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		v, err := s.RunDefault()
		r.NoError(t, err)
		r.True(t, v.ToBoolean())
		s.Close()
	})
	t.Run("typescript", func(t *testing.T) {
		t.Parallel()

		s, err := jstest.New(jstest.WithPathSource("test.ts", `const msg: string = 'Hello World'; export default function() { return msg }`), js.WithHost(&enginetest.Host{}))
		r.NoError(t, err)
		v, err := s.RunDefault()
		r.NoError(t, err)
		r.Equal(t, "Hello World", v.String())
		s.Close()
	})
}
