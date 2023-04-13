package common

import "net/http"

type EventEmitter interface {
	Emit(event string, args ...interface{}) []*Action
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

	On(event string, do func(args ...interface{}) (bool, error), tags map[string]string)

	KafkaClient() KafkaClient
	HttpClient() HttpClient

	Name() string
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type KafkaClient interface {
	Produce(args *KafkaProduceArgs) (interface{}, interface{}, error)
}

type KafkaProduceArgs struct {
	Cluster   string
	Topic     string
	Partition int
	Key       interface{}
	Value     interface{}
	Headers   map[string]interface{}
	Timeout   int
}

type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type Action struct {
	Duration int64             `json:duration`
	Tags     map[string]string `json:"tags"`
}

func NewJobOptions() JobOptions {
	return JobOptions{
		Tags:                    map[string]string{},
		Times:                   -1,
		RunFirstTimeImmediately: true,
	}
}
