package kafka

import (
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
)

type Client interface {
	Produce(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error)
}

type Module struct {
	host   common.Host
	rt     *goja.Runtime
	client Client
}

type produceParams struct {
	Cluster   string
	Topic     string
	Partition int
	Key       interface{}
	Value     interface{}
	Headers   map[string]interface{}
	Timeout   int
}

type ProduceResult struct {
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
	Error string      `json:"error"`
}

func New(host common.Host, rt *goja.Runtime) interface{} {
	return &Module{host: host, rt: rt, client: host.KafkaClient()}
}

func (m *Module) Produce(v goja.Value) interface{} {
	params := mapParams(v, m.rt)
	r := &ProduceResult{}
	var err error
	r.Key, r.Value, err = m.client.Produce(params.Cluster, params.Topic, params.Partition, params.Key, params.Value, params.Headers)
	if err != nil {
		log.Errorf("js error: %v in %v", err, m.host.Name())
		r.Error = err.Error()
	}
	return r
}

func mapParams(args goja.Value, rt *goja.Runtime) *produceParams {
	opt := &produceParams{Timeout: 30, Partition: -1}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(rt)
		for _, k := range params.Keys() {
			switch k {
			case "cluster":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Cluster = tagsV.String()
			case "topic":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Topic = tagsV.String()
			case "partition":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Partition = int(tagsV.ToInteger())
			case "key":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Key = tagsV.Export()
			case "value":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Value = tagsV.Export()
			case "headers":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Headers = make(map[string]interface{})
				tagsO := tagsV.ToObject(rt)
				for _, key := range tagsO.Keys() {
					opt.Headers[key] = tagsO.Get(key).String()
				}
			case "timeout":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Timeout = int(tagsV.ToInteger())
			}
		}
	}
	return opt
}
