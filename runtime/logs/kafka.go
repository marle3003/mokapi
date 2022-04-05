package logs

import "time"

type KafkaMessage struct {
	Key     string
	Message string
	Time    time.Time
}

func NewKafkaLog(key, message string) *KafkaMessage {
	return &KafkaMessage{
		Key:     key,
		Message: message,
		Time:    time.Now(),
	}
}
