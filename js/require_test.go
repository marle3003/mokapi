package js

import (
	"github.com/dop251/goja"
	"mokapi/js/common"
	"mokapi/test"
	"testing"
)

func TestRequire(t *testing.T) {
	host := struct{ common.Host }{}

	t.Parallel()
	t.Run("import", func(t *testing.T) {
		t.Parallel()
		s, err := New("test", `import {sleep} from 'mokapi'; export let _sleep = sleep; sleep(12); export default function() {}`, host)
		test.Ok(t, err)

		exports := s.runtime.Get("exports").ToObject(s.runtime)
		_, ok := goja.AssertFunction(exports.Get("_sleep"))
		test.Assert(t, ok, "sleep is not a function")
	})
}
