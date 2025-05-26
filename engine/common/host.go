package common

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/schema/json/generator"
	"net/http"
	"strings"
	"time"
)

type EventEmitter interface {
	Emit(event string, args ...interface{}) []*Action
}

type Script interface {
	Run() error
	Close()
	CanClose() bool
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
	HttpClient(HttpClientOptions) HttpClient

	Name() string

	FindFakerNode(name string) *generator.Node
	AddCleanupFunc(f func())

	Lock()
	Unlock()
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	IsLevelEnabled(level string) bool
}

type KafkaClient interface {
	Produce(args *KafkaProduceArgs) (*KafkaProduceResult, error)
}

type KafkaProduceArgs struct {
	Cluster  string
	Topic    string
	Messages []KafkaMessage
	Timeout  int
	Retry    KafkaProduceRetry
}

type KafkaMessage struct {
	Key       interface{}
	Value     []byte
	Data      interface{}
	Headers   map[string]interface{}
	Partition int
}

type KafkaProduceRetry struct {
	MaxRetryTime     time.Duration
	InitialRetryTime time.Duration
	Factor           int
	Retries          int
}

type KafkaProduceResult struct {
	Cluster  string
	Topic    string
	Messages []KafkaMessageResult
}

type KafkaMessageResult struct {
	Key       string
	Value     string
	Offset    int64
	Headers   map[string]string
	Partition int
}

type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type HttpClientOptions struct {
	MaxRedirects int
}

type Action struct {
	Duration   int64             `json:"duration"`
	Tags       map[string]string `json:"tags"`
	Parameters []any             `json:"parameters"`
	Logs       []Log             `json:"logs"`
	Error      *Error            `json:"error"`
}

type Log struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type JobExecution struct {
	Schedule string    `json:"schedule"`
	MaxRuns  int       `json:"maxRuns"`
	Runs     int       `json:"runs"`
	NextRun  time.Time `json:"nextRun"`

	Duration int64             `json:"duration"`
	Tags     map[string]string `json:"tags"`
	Logs     []Log             `json:"logs"`
	Error    *Error            `json:"error"`
}

type Error struct {
	Message string `json:"message"`
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

func (a *Action) AppendLog(level, message string) {
	a.Logs = append(a.Logs, Log{Level: level, Message: message})
}

func (e *JobExecution) AppendLog(level, message string) {
	e.Logs = append(e.Logs, Log{Level: level, Message: message})
}

type FakerNode interface {
	Name() string
	Fake(r *generator.Request) (interface{}, error)
}
