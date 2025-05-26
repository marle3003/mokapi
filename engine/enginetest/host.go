package enginetest

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/schema/json/generator"
	"net/http"
	"net/url"
	"sync"
)

type Host struct {
	CleanupFuncs []func()

	OpenFileFunc       func(file, hint string) (string, string, error)
	OpenFunc           func(file, hint string) (*dynamic.Config, error)
	InfoFunc           func(args ...interface{})
	WarnFunc           func(args ...interface{})
	ErrorFunc          func(args ...interface{})
	DebugFunc          func(args ...interface{})
	IsLevelEnabledFunc func(level string) bool
	HttpClientTest     *HttpClient
	KafkaClientTest    *KafkaClient
	EveryFunc          func(every string, do func(), opt common.JobOptions)
	CronFunc           func(every string, do func(), opt common.JobOptions)
	OnFunc             func(event string, do func(args ...interface{}) (bool, error), tags map[string]string)
	FindFakerNodeFunc  func(name string) *generator.Node
	m                  sync.Mutex
}

type HttpClient struct {
	LastRequest *http.Request
	DoFunc      func(request *http.Request) (*http.Response, error)
}

type KafkaClient struct {
	ProduceFunc func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error)
}

func (h *Host) Info(args ...interface{}) {
	if h.InfoFunc != nil {
		h.InfoFunc(args...)
	}
}

func (h *Host) Warn(args ...interface{}) {
	if h.WarnFunc != nil {
		h.WarnFunc(args...)
	}
}

func (h *Host) Error(args ...interface{}) {
	if h.ErrorFunc != nil {
		h.ErrorFunc(args...)
	}
}

func (h *Host) Debug(args ...interface{}) {
	if h.DebugFunc != nil {
		h.DebugFunc(args...)
	}
}

func (h *Host) IsLevelEnabled(level string) bool {
	if h.IsLevelEnabledFunc == nil {
		return true
	}
	return h.IsLevelEnabledFunc(level)
}

func (h *Host) OpenFile(file, hint string) (*dynamic.Config, error) {
	if h.OpenFileFunc != nil {
		p, src, err := h.OpenFileFunc(file, hint)
		if err != nil {
			return nil, err
		}
		return &dynamic.Config{Raw: []byte(src), Info: dynamic.ConfigInfo{Url: mustParse(p)}}, nil
	}
	if h.OpenFunc != nil {
		return h.OpenFunc(file, hint)
	}
	return nil, fmt.Errorf("file %v not found (hint: %v)", file, hint)
}

func (h *Host) Every(every string, do func(), opt common.JobOptions) (int, error) {
	if h.EveryFunc != nil {
		h.EveryFunc(every, do, opt)
	}
	return 0, nil
}

func (h *Host) Cron(expr string, do func(), opt common.JobOptions) (int, error) {
	if h.CronFunc != nil {
		h.CronFunc(expr, do, opt)
	}
	return 0, nil
}

func (h *Host) On(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
	if h.OnFunc != nil {
		h.OnFunc(event, do, tags)
	}
}

func (h *Host) Cancel(jobId int) error {
	return nil
}

func (h *Host) Name() string {
	return "test host"
}

func (h *Host) FindFakerNode(name string) *generator.Node {
	if h.FindFakerNodeFunc != nil {
		return h.FindFakerNodeFunc(name)
	}
	return nil
}

func (h *Host) Lock() {
	h.m.Lock()
}

func (h *Host) Unlock() {
	h.m.Unlock()
}

func (h *Host) HttpClient(opts common.HttpClientOptions) common.HttpClient {
	return h.HttpClientTest
}

func (h *Host) KafkaClient() common.KafkaClient {
	return h.KafkaClientTest
}

func (c *HttpClient) Do(request *http.Request) (*http.Response, error) {
	c.LastRequest = request
	if c.DoFunc != nil {
		return c.DoFunc(request)
	}
	return &http.Response{}, nil
}

func (c *KafkaClient) Produce(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
	if c.ProduceFunc != nil {
		return c.ProduceFunc(args)
	}
	return nil, nil
}

func (h *Host) AddCleanupFunc(f func()) {
	h.CleanupFuncs = append(h.CleanupFuncs, f)
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
