package common

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/ldap"
	"mokapi/schema/json/generator"
	"mokapi/smtp"
	"net/http"
	"strings"
	"time"
)

type EventEmitter interface {
	HttpEventEmitter
	KafkaEventEmitter
	MqttEventEmitter
	WebsocketEventEmitter
	MailEventEmitter
	LdapEventEmitter
}

type HttpEventEmitter interface {
	EmitHttp(request *HttpEventRequest, response *HttpEventResponse) []*Action
}

type KafkaEventEmitter interface {
	EmitKafka(record *KafkaEventRecord) []*Action
}

type MqttEventEmitter interface {
	EmitMqtt(message *MqttMessageEvent) []*Action
}

type WebsocketEventEmitter interface {
	EmitWebsocketConnect(event *WebsocketConnectEvent) []*Action
	EmitWebsocketMessage(event *WebsocketMessageEvent) []*Action
	EmitWebsocketClose(event *WebsocketCloseEvent) []*Action
}

type MailEventEmitter interface {
	EmitSmtp(message *smtp.Message, status *smtp.Status) []*Action
}

type LdapEventEmitter interface {
	EmitLdap(request *ldap.SearchRequest, response *ldap.SearchResponse) []*Action
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

type EventArgs struct {
	Tags     map[string]string
	Priority int
}

type Host interface {
	Logger
	SetEventLogger(func(level, message string))

	Every(every string, do func(), opt JobOptions) (int, error)
	Cron(expr string, do func(), opt JobOptions) (int, error)
	Cancel(jobId int) error

	OpenFile(file string, hint string) (*dynamic.Config, error)

	OnHttp(filter HttpFilter, do EventHandler, args EventArgs)
	OnKafka(filter KafkaFilter, do EventHandler, args EventArgs)
	OnMqtt(filter MqttFilter, do EventHandler, args EventArgs)
	OnWebsocket(filter WebsocketFilter, do EventHandler, args EventArgs)
	OnMail(filter MailFilter, do EventHandler, args EventArgs)
	OnLdap(filter LdapFilter, do EventHandler, args EventArgs)

	Webhook(name string, url string, args WebhookArgs) (*WebhookResponse, error)

	KafkaClient() KafkaClient
	MqttClient() MqttClient
	HttpClient(HttpClientOptions) HttpClient

	Name() string

	FindFakerNode(name string) *generator.Node
	AddCleanupFunc(f func())

	Lock()
	Unlock()

	Store() Store

	Cwd() string
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
	Cluster    string
	Topic      string
	Messages   []KafkaMessage
	Timeout    int
	Retry      RetryArgs
	ClientId   string
	ScriptFile string
}

type KafkaMessage struct {
	Key       interface{}
	Value     []byte
	Data      interface{}
	Headers   map[string]string
	Partition int
}

type RetryArgs struct {
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
	Timeout      time.Duration
	Insecure     bool
}

type MqttClient interface {
	Publish(args *MqttPublishArgs) (*MqttPublishResult, error)
}

type MqttPublishArgs struct {
	Cluster    string
	Topic      string
	Value      string
	Retain     bool
	Timeout    int
	Retry      RetryArgs
	ClientId   string
	ScriptFile string
}

type MqttPublishResult struct {
	Cluster string
	Topic   string
	Value   string
}

type Action struct {
	Duration   int64             `json:"duration"`
	Tags       map[string]string `json:"tags"`
	Parameters []string          `json:"parameters"`
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

func (e *JobExecution) Title() string {
	if e.Error != nil {
		return fmt.Sprintf("Error in: %v", e.Tags["name"])
	}
	return e.Tags["name"]
}

func (e *JobExecution) Domain() string {
	return "Job Execution"
}

type FakerNode interface {
	Name() string
	Fake(r *generator.Request) (interface{}, error)
}

type Store interface {
	Get(string) any
	Set(string, any)
	Has(string) bool
	Delete(string)
	Clear()
	Update(key string, fn func(v any) any) any
	Keys() []string
	Namespace(name string) Store
}

type EventHandler func(ctx *EventContext) (bool, error)

type EventContext struct {
	EventLogger func(level, message string)
	Args        []any
}

type WebhookArgs struct {
	Method   string         `json:"method"`
	Data     any            `json:"data"`
	Body     string         `json:"body"`
	Headers  map[string]any `json:"headers"`
	Api      string         `json:"api"`
	Timeout  time.Duration  `json:"timeout"`
	Insecure bool           `json:"insecure"`
}

type WebhookResponse struct {
	StatusCode int            `json:"statusCode"`
	Data       any            `json:"data"`
	Headers    map[string]any `json:"headers"`
}
