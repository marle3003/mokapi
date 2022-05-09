package logs

import "time"

type KafkaMessage struct {
	Offset  int64
	Key     string
	Message string
	Time    time.Time
}

func NewKafkaLog(offset int64, key, message string, time time.Time) *KafkaMessage {
	return &KafkaMessage{
		Offset:  offset,
		Key:     key,
		Message: message,
		Time:    time,
	}
}
