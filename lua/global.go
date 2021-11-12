package lua

import "time"

func sleep(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}
