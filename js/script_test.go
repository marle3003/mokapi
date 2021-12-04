package js

import (
	"fmt"
	"github.com/dop251/goja"
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
		_, err := New("", "", host)
		test.Equals(t, fmt.Errorf("no exported functions in script"), err)
	})
	t.Run("null", func(t *testing.T) {
		t.Parallel()
		_, err := New("", "exports = null", host)
		test.Equals(t, fmt.Errorf("export must be an object"), err)
	})
	t.Run("emptyFunction", func(t *testing.T) {
		t.Parallel()
		_, err := New("test", `export default function() {}`, host)
		test.Ok(t, err)
	})
	t.Run("alert", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {console.log("foo")}`, host)
		test.Ok(t, err)
		_, err = s.exports["default"](goja.Undefined())
		test.Ok(t, err)
	})
	t.Run("returnValueFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {return 2}`, host)
		test.Ok(t, err)
		v, err := s.exports["default"](goja.Undefined())
		test.Ok(t, err)
		test.Equals(t, int64(2), v.ToInteger())
	})
	t.Run("customFunction", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `function custom() {return 2}; export {custom}`, host)
		test.Ok(t, err)
		v, err := s.exports["custom"](goja.Undefined())
		test.Ok(t, err)
		test.Equals(t, int64(2), v.ToInteger())
	})
	t.Run("interrupt", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `export default function() {while(true) {}}`, host)
		test.Ok(t, err)
		go func() {
			_, err := s.exports["default"](goja.Undefined())
			iErr := err.(*goja.InterruptedError)
			test.Assert(t, strings.HasPrefix(iErr.String(), "closing"), "closed execution")
		}()
		s.Close()
	})
}

type testHost struct {
	common.Host
}

func (th *testHost) Info(args ...interface{}) {

}
