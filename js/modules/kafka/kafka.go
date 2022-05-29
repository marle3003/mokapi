package kafka

import (
	"github.com/dop251/goja"
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

type produceOptions struct {
	Cluster   string
	Topic     string
	Partition int
	Key       interface{}
	Value     interface{}
	Headers   map[string]interface{}
	Timeout   int
}

func New(host common.Host, rt *goja.Runtime) interface{} {
	return &Module{host: host, rt: rt, client: host.KafkaClient()}
}

func (m *Module) Produce(args goja.Value) (interface{}, interface{}, error) {
	opt := m.mapOption(args)
	return m.client.Produce(opt.Cluster, opt.Topic, opt.Partition, opt.Key, opt.Value, opt.Headers)
}

func (m *Module) mapOption(args goja.Value) *produceOptions {
	opt := &produceOptions{Timeout: 30, Partition: -1}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
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
				opt.Key = tagsV.ToObject(m.rt)
			case "value":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Key = tagsV.ToObject(m.rt)
			case "headers":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				opt.Headers = make(map[string]interface{})
				tagsO := tagsV.ToObject(m.rt)
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
