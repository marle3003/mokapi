package common

import "net/http"

type EventEmitter interface {
	Emit(event string, args ...interface{})
}

type Script interface {
	Run() error
	Close()
}

type JobOptions struct {
	Times                   int
	RunFirstTimeImmediately bool
	Tags                    map[string]string
}

type Host interface {
	Logger
	Every(every string, do func(), opt JobOptions) (int, error)
	Cron(expr string, do func(), opt JobOptions) (int, error)
	Cancel(jobId int) error

	OpenFile(file string, hint string) (string, string, error)
	OpenScript(file string, hint string) (string, string, error)

	On(event string, do func(args ...interface{}) (bool, error), tags map[string]string)

	KafkaClient() KafkaClient
	HttpClient() HttpClient
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type KafkaClient interface {
	Produce(cluster string, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error)
}

type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

func NewJobOptions() JobOptions {
	return JobOptions{
		Times:                   -1,
		RunFirstTimeImmediately: true,
	}
}
