package eventloop_test

import (
	"mokapi/engine/enginetest"
	"mokapi/js/eventloop"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestEventLoop(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, loop *eventloop.EventLoop)
	}{
		{
			name: "setTimeout",
			test: func(t *testing.T, loop *eventloop.EventLoop) {
				_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
					_ = vm.Set("now", time.Now)
					return nil, nil
				})
				require.NoError(t, err)
				startTime := time.Now()

				loop.Run(func(vm *goja.Runtime) {
					_, err := vm.RunString(`
						var calledAt;
						setTimeout(() => {
							calledAt = now()	
						}, 1000)`)
					require.NoError(t, err)
				})

				time.Sleep(1500 * time.Millisecond)
				var calledAt time.Time
				_, err = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
					v := vm.Get("calledAt")
					err := vm.ExportTo(v, &calledAt)
					return v, err
				})
				require.NoError(t, err)

				require.False(t, calledAt.IsZero(), "calledAt should not be zero")
				require.Greater(t, calledAt.Sub(startTime), time.Second, "code should wait for a second")
			},
		},
		{
			name: "clearTimeout",
			test: func(t *testing.T, loop *eventloop.EventLoop) {
				loop.Run(func(vm *goja.Runtime) {
					_, err := vm.RunString(`
						const id = setTimeout(() => {
							throw new Error("timer should not run")
						}, 1000)`)
					require.NoError(t, err)
				})

				time.Sleep(500 * time.Millisecond)
				_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
					return vm.RunString(`
						clearTimeout(id)`)

				})
				require.NoError(t, err)
				require.False(t, loop.HasJobs())
			},
		},
		{
			name: "setInterval",
			test: func(t *testing.T, loop *eventloop.EventLoop) {
				loop.Run(func(vm *goja.Runtime) {
					_, err := vm.RunString(`
						var counter = 0;
						setInterval(() => {
							counter++
						}, 100)`)
					require.NoError(t, err)
				})

				time.Sleep(500 * time.Millisecond)
				var counter int64
				_, _ = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
					v := vm.Get("counter")
					counter = v.ToInteger()
					return nil, nil
				})

				require.Greater(t, counter, int64(3))
			},
		},
		{
			name: "clearInterval",
			test: func(t *testing.T, loop *eventloop.EventLoop) {
				loop.Run(func(vm *goja.Runtime) {
					_, err := vm.RunString(`
						var counter = 0;
						const i = setInterval(() => {
							counter++
						}, 100)`)
					require.NoError(t, err)
				})

				time.Sleep(300 * time.Millisecond)
				_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
					return vm.RunString("clearInterval(i)")
				})
				require.NoError(t, err)

				var counter int64
				_, _ = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
					v := vm.Get("counter")
					counter = v.ToInteger()
					return nil, nil
				})

				require.LessOrEqual(t, counter, int64(3))
				require.False(t, loop.HasJobs())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := &enginetest.Host{}
			vm := goja.New()
			loop := eventloop.New(vm, h)
			loop.StartLoop()
			defer loop.Stop()

			tc.test(t, loop)
		})
	}
}
