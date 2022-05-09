package common

type EventEmitter interface {
	Emit(event string, args ...interface{})
}

type Host interface {
	Logger
	Every(every string, do func(), times int, tags map[string]string) (int, error)
	Cron(expr string, do func(), times int, tags map[string]string) (int, error)
	Cancel(jobId int) error

	OpenFile(file string) (string, error)

	On(event string, do func(args ...interface{}) (bool, error), tags map[string]string)

	KafkaClient() KafkaClient
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type KafkaClient interface {
	Produce(cluster string, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error)
}
