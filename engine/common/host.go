package common

import (
	"fmt"
	"mokapi/config/dynamic"
	"net/http"
	"strings"
)

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

	OpenFile(file string, hint string) (*dynamic.Config, error)

	On(event string, do func(args ...interface{}) (bool, error), tags map[string]string)

	KafkaClient() KafkaClient
	HttpClient() HttpClient

	Name() string
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
}

type KafkaClient interface {
	Produce(args *KafkaProduceArgs) (*KafkaProduceResult, error)
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

type KafkaProduceResult struct {
	Cluster   string
	Topic     string
	Partition int
	Offset    int64
	Key       string
	Value     string
	Headers   map[string]string
}

type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type Action struct {
	Duration int64             `json:"duration"`
	Tags     map[string]string `json:"tags"`
}

func NewJobOptions() JobOptions {
	return JobOptions{
		Tags:                    map[string]string{},
		Times:                   -1,
		RunFirstTimeImmediately: true,
	}
}

func (a *Action) String() string {
	var sb strings.Builder
	for k, v := range a.Tags {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v=%v", k, v))
	}
	return sb.String()
}
