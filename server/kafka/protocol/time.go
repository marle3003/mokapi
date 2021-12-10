package protocol

import "time"

const (
	Earliest int64 = -2 // first offset
	Latest   int64 = -1 // last offset
)

func Timestamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano() / int64(time.Millisecond)
}

func (a Attributes) Compression() int8 {
	return int8(a & 7)
}

func toTime(i int64) time.Time {
	return time.Unix(i/1000, 0)
}
