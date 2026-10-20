package try

import (
	"testing"
	"time"
)

func MustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func WaitFor(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)

	for {
		if check() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("waitFor deadline")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
