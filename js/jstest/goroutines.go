package jstest

import (
	"runtime"
	"testing"
	"time"
)

// RequireGoroutinesAtMost fails the test if goroutines are still running above the given
// count after two seconds. It polls instead of using require.Eventually, because Eventually
// runs its condition in goroutines of its own and would count itself.
func RequireGoroutinesAtMost(t *testing.T, count int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > count {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines leaked: %d running, expected at most %d", runtime.NumGoroutine(), count)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
