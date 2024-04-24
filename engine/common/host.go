package common

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/json/generator"
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
	Times                 int
	SkipImmediateFirstRun bool
	Tags                  map[string]string
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

	FindFakerTree(name string) FakerTree

	Lock()
	Unlock()
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
	Cluster  string
	Topic    string
	Messages []KafkaMessage
	Timeout  int
}

type KafkaMessage struct {
	Key       interface{}
	Value     []byte
	Data      interface{}
	Headers   map[string]interface{}
	Partition int
}

type KafkaProduceResult struct {
	Cluster  string
	Topic    string
	Messages []KafkaProducedMessage
}

type KafkaProducedMessage struct {
	Key       string
	Value     string
	Offset    int64
	Headers   map[string]string
	Partition int
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
		Tags:                  map[string]string{},
		Times:                 -1,
		SkipImmediateFirstRun: false,
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

type FakerTree interface {
	Name() string
	Test(r *generator.Request) bool
	Fake(r *generator.Request) (interface{}, error)

	Append(tree FakerNode)
	Insert(index int, tree FakerNode) error
	RemoveAt(index int) error
	Remove(name string) error
}

type FakerNode interface {
	Name() string
	Test(r *generator.Request) bool
	Fake(r *generator.Request) (interface{}, error)
}
