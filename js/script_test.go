package js

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/test"
	"strings"
	"testing"
)

func TestScript(t *testing.T) {
	host := &testHost{}

	t.Parallel()
	t.Run("blank", func(t *testing.T) {
		t.Parallel()
		s, err := New("", "", host)
		test.Ok(t, err)
		err = s.Run()
		test.Equals(t, fmt.Errorf("no exported functions in script"), err)
	})
	t.Run("null", func(t *testing.T) {
		t.Parallel()
		s, err := New("", "exports = null", host)
		test.Ok(t, err)
		err = s.Run()
		test.Equals(t, fmt.Errorf("export must be an object"), err)
	})
	t.Run("emptyFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {}`, host)
		test.Ok(t, err)
		test.Ok(t, s.Run())
	})
	t.Run("console.log", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {console.log("foo")}`, host)
		test.Ok(t, err)
		test.Ok(t, s.Run())
	})
	t.Run("returnValueFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {return 2}`, host)
		test.Ok(t, err)
		test.Ok(t, s.Run())
		v, err := s.exports["default"](goja.Undefined())
		test.Ok(t, err)
		test.Equals(t, int64(2), v.ToInteger())
	})
	t.Run("customFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `function custom() {return 2}; export {custom}`, host)
		test.Ok(t, err)
		test.Ok(t, s.Run())
		v, err := s.exports["custom"](goja.Undefined())
		test.Ok(t, err)
		test.Equals(t, int64(2), v.ToInteger())
	})
	t.Run("interrupt", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {while(true) {}}`, host)
		test.Ok(t, err)
		ch := make(chan bool)
		go func() {
			ch <- true
			err := s.Run()
			iErr := err.(*goja.InterruptedError)
			test.Assert(t, strings.HasPrefix(iErr.String(), "closing"), fmt.Sprintf("error prefix expected closing but got: %v", iErr.String()))
		}()

		_ = <-ch
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
import g from 'generator'
export default function() {
  var s = g.new({type: 'string'})
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
}

func (th *testHost) Info(args ...interface{}) {

}
