package js_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
	"time"
)

func TestTypeScript(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "setTimeout",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithPathSource(
					"test.ts",
					`
var calledAt;
export default function() {
	setTimeout(() => { calledAt = now() }, 1000)
}
`),
					js.WithHost(host))
				r.NoError(t, err)
				defer s.Close()

				err = s.RunFunc(func(vm *goja.Runtime) {
					_ = vm.Set("now", time.Now)
				})
				r.NoError(t, err)

				startTime := time.Now()
				err = s.Run()
				r.NoError(t, err)

				time.Sleep(2 * time.Second)
				var calledAt time.Time
				err = s.RunFunc(func(vm *goja.Runtime) {
					v := vm.Get("calledAt")
					err = vm.ExportTo(v, &calledAt)
					r.NoError(t, err)
				})
				time.Sleep(1 * time.Second)
				r.False(t, calledAt.IsZero(), "calledAt should not be zero")
				r.Greater(t, calledAt.Sub(startTime), time.Second, "code should wait for a second")
			},
		},
		{
			name: "async",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithPathSource(
					"test.ts",
					`
let msg1;
let msg2;
export default function() {
	Promise.all([getMessage(), getMessage()]).then((values) => {
		msg1 = values[0]
		msg2 = values[1]
	})
}
let getMessage = async () => {
	return new Promise(async (resolve, reject) => {
	  setTimeout(() => {
		resolve('foo');
	  }, 200);
	});
}
`),
					js.WithHost(host))
				r.NoError(t, err)
				defer s.Close()

				err = s.Run()
				r.NoError(t, err)

				time.Sleep(1 * time.Second)
				var msg1 string
				var msg2 string
				err = s.RunFunc(func(vm *goja.Runtime) {
					v := vm.Get("msg1")
					err = vm.ExportTo(v, &msg1)
					r.NoError(t, err)
					v = vm.Get("msg2")
					err = vm.ExportTo(v, &msg2)
					r.NoError(t, err)
				})
				time.Sleep(2 * time.Second)
				r.Equal(t, "foo", msg1)
				r.Equal(t, "foo", msg2)
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
