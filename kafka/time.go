package kafka

import "time"

const (
	// Earliest the earliest available offset
	Earliest int64 = -2
	// Latest the offset of the next coming message
	Latest int64 = -1
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

func ToTime(i int64) time.Time {
	return time.Unix(i/1000, (i%1000)*int64(time.Millisecond))
}
