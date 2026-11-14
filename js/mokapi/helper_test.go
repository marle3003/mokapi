package mokapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

// The goja runtime is not goroutine-safe: while the event loop runs, tests must not touch
// the runtime from the test goroutine. These helpers run everything on the loop.

// evalOnLoop runs the script on the event loop and returns the exported result.
func evalOnLoop(loop *eventloop.EventLoop, script string) (any, error) {
	var exported any
	_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		v, err := vm.RunString(script)
		if err == nil && v != nil {
			exported = v.Export()
		}
		return v, err
	})
	return exported, err
}

// awaitPromise waits until the promise stored in the global variable name is settled and
// returns its state. It fails the test if the promise is still pending after the deadline.
func awaitPromise(t *testing.T, loop *eventloop.EventLoop, name string) goja.PromiseState {
	t.Helper()
	state, _ := awaitPromiseResult(t, loop, name)
	return state
}

// awaitPromiseResult is like awaitPromise and also returns the exported value or reason.
func awaitPromiseResult(t *testing.T, loop *eventloop.EventLoop, name string) (goja.PromiseState, any) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var state goja.PromiseState
		var result any
		_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
			p := vm.Get(name).Export().(*goja.Promise)
			state = p.State()
			if state != goja.PromiseStatePending {
				result = p.Result().Export()
			}
			return nil, nil
		})
		r.NoError(t, err)
		if state != goja.PromiseStatePending {
			return state, result
		}
		if time.Now().After(deadline) {
			t.Fatalf("promise %q was not settled", name)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// scriptFieldNameMapper names fields and methods as a script sees them in Mokapi: the json tag
// or the uncapitalized Go name (e.g. MqttPublishResult.Api is `api`). goja.TagFieldNameMapper
// keeps the Go name of fields without a tag, which would hide naming bugs.
type scriptFieldNameMapper struct{}

func (scriptFieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	if tag, _, _ := strings.Cut(f.Tag.Get("json"), ","); tag != "" {
		return tag
	}
	return uncapitalize(f.Name)
}

func (scriptFieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return uncapitalize(m.Name)
}

func uncapitalize(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

// newModuleLoop returns a started event loop with the mokapi module enabled. The loop is
// stopped when the test ends.
func newModuleLoop(t *testing.T) (*eventloop.EventLoop, *enginetest.Host) {
	t.Helper()
	reg, err := require.NewRegistry(nil)
	r.NoError(t, err)
	reg.RegisterNativeModule("mokapi", mokapi.Require)

	vm := goja.New()
	vm.SetFieldNameMapper(scriptFieldNameMapper{})
	host := &enginetest.Host{}
	loop := eventloop.New(vm, host)
	loop.StartLoop()
	t.Cleanup(loop.Stop)
	js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
	reg.Enable(vm)
	return loop, host
}
