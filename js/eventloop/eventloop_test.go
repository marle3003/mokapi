package eventloop_test

import (
	"errors"
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

func TestEventLoop_NestedRunSync_NoDeadlock(t *testing.T) {
	vm := goja.New()
	loop := eventloop.New(vm, &enginetest.Host{})
	loop.StartLoop()
	defer loop.Stop()

	done := make(chan struct{})

	go func() {
		_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
			return loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
				return vm.ToValue("inner result"), nil
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
		// ok, kein Deadlock
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock: nested RunSync did not return in time")
	}
}

// Timers fire on their own goroutines while the loop keeps running scripts. The job counter
// and the running flag are shared between these goroutines, so this test is meant to run
// with `go test -race`.
func TestEventLoop_TimersFiredConcurrently(t *testing.T) {
	vm := goja.New()
	loop := eventloop.New(vm, &enginetest.Host{})
	loop.StartLoop()
	defer loop.Stop()

	_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		return vm.RunString(`
			var fired = 0
			for (let i = 0; i < 50; i++) {
				setTimeout(() => { fired++ }, 1)
			}
			var ticks = 0
			var id = setInterval(() => {
				if (++ticks === 5) { clearInterval(id) }
			}, 1)
		`)
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		var done bool
		_, err := loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
			done = vm.Get("fired").ToInteger() == 50 && vm.Get("ticks").ToInteger() >= 5
			return nil, nil
		})
		return err == nil && done && !loop.HasJobs()
	}, 5*time.Second, 5*time.Millisecond)
}

// The caller of RunAsync must not read the promise while the loop settles it. A single
// timer is not enough to show this, because the caller only wakes up after the settling
// job. Many timers make other jobs run while the promise is still pending.
func TestEventLoop_RunAsync_PromiseSettledLater(t *testing.T) {
	testcases := []struct {
		name    string
		settle  string
		value   string
		wantErr string
	}{
		{name: "resolved", settle: "resolve('foo')", value: "foo"},
		{name: "rejected", settle: "reject('bar')", wantErr: "bar"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			loop := eventloop.New(vm, &enginetest.Host{})
			loop.StartLoop()
			defer loop.Stop()

			v, err := loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
				return vm.RunString(`
					new Promise((resolve, reject) => {
						for (let i = 0; i < 100; i++) {
							setTimeout(() => {}, 10)
						}
						setTimeout(() => { ` + tc.settle + ` }, 10)
					})`)
			}, &eventloop.JobContext{})

			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.value, v.Export())
		})
	}
}

func TestEventLoop_RunAsync(t *testing.T) {
	testcases := []struct {
		name    string
		job     eventloop.JobFunc
		want    interface{}
		wantErr string
	}{
		{
			name: "promise already resolved",
			job:  script(`Promise.resolve('foo')`),
			want: "foo",
		},
		{
			name:    "promise already rejected",
			job:     script(`Promise.reject('bar')`),
			wantErr: "bar",
		},
		{
			name:    "rejected with Error object",
			job:     script(`Promise.reject(new Error('x'))`),
			wantErr: "Error: x",
		},
		{
			name: "async function awaiting setTimeout",
			job: script(`
				(async () => {
					await new Promise(resolve => setTimeout(resolve, 10))
					return 'done'
				})()`),
			want: "done",
		},
		{
			name:    "fn returns error",
			job:     func(vm *goja.Runtime) (goja.Value, error) { return nil, errors.New("failed") },
			wantErr: "failed",
		},
		{
			name:    "fn panics",
			job:     func(vm *goja.Runtime) (goja.Value, error) { panic("boom") },
			wantErr: "boom",
		},
		{
			name: "non-promise result",
			job:  script(`'foo'`),
			want: "foo",
		},
		{
			name: "nil result",
			job:  func(vm *goja.Runtime) (goja.Value, error) { return nil, nil },
			want: nil,
		},
		{
			name: "undefined result",
			job:  script(`undefined`),
			want: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop := startLoop(t)

			v, err := runAsyncWithTimeout(t, loop, tc.job)

			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, v)
			require.Equal(t, tc.want, v.Export())
		})
	}
}

func script(code string) eventloop.JobFunc {
	return func(vm *goja.Runtime) (goja.Value, error) { return vm.RunString(code) }
}

func startLoop(t *testing.T) *eventloop.EventLoop {
	loop := eventloop.New(goja.New(), &enginetest.Host{})
	loop.StartLoop()
	t.Cleanup(loop.Stop)
	return loop
}

// runAsyncWithTimeout makes a hanging RunAsync fail the test instead of blocking it.
func runAsyncWithTimeout(t *testing.T, loop *eventloop.EventLoop, job eventloop.JobFunc) (goja.Value, error) {
	type outcome struct {
		value goja.Value
		err   error
	}
	done := make(chan outcome, 1)
	go func() {
		v, err := loop.RunAsync(job, &eventloop.JobContext{})
		done <- outcome{v, err}
	}()

	select {
	case o := <-done:
		return o.value, o.err
	case <-time.After(2 * time.Second):
		t.Fatal("RunAsync did not return in time")
		return nil, nil
	}
}

func TestEventLoop_RunAsync_StopWhilePending(t *testing.T) {
	loop := startLoop(t)

	time.AfterFunc(100*time.Millisecond, loop.Stop)
	_, err := runAsyncWithTimeout(t, loop, script(`new Promise(() => {})`))

	require.ErrorIs(t, err, eventloop.ErrStopped)
}

func TestEventLoop_RunAsync_AfterStop(t *testing.T) {
	loop := startLoop(t)
	loop.Stop()
	time.Sleep(50 * time.Millisecond)

	_, err := runAsyncWithTimeout(t, loop, script(`'foo'`))

	require.ErrorIs(t, err, eventloop.ErrStopped)
}

func TestEventLoop_HasJobs_WhilePending(t *testing.T) {
	loop := startLoop(t)

	_, err := loop.RunSync(script(`
		var settle
		var promise = new Promise(resolve => { settle = resolve })
		undefined`))
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		_, err := loop.RunAsync(script(`promise`), &eventloop.JobContext{})
		done <- err
	}()

	require.Eventually(t, loop.HasJobs, 2*time.Second, time.Millisecond, "pending promise should keep the loop busy")

	_, err = loop.RunSync(script(`settle('foo')`))
	require.NoError(t, err)

	require.NoError(t, <-done)
	require.Eventually(t, func() bool { return !loop.HasJobs() }, 2*time.Second, time.Millisecond)
}

func TestEventLoop_Stop_Twice(t *testing.T) {
	loop := startLoop(t)

	done := make(chan struct{})
	go func() {
		loop.Stop()
		loop.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("second Stop did not return in time")
	}
}
