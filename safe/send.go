package safe

import "context"

// Send delivers v on ch unless ctx is done first. It reports whether v was delivered.
// Producers use it so that they never block on a consumer that has stopped reading at shutdown.
func Send[T any](ctx context.Context, ch chan<- T, v T) bool {
	select {
	case ch <- v:
		return true
	case <-ctx.Done():
		return false
	}
}
