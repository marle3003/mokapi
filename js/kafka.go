package js

import (
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
)

type kafkaModule struct {
	host   common.Host
	rt     *goja.Runtime
	client common.KafkaClient
}

type ProduceResult struct {
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
	Error string      `json:"error"`
}

func newKafka(host common.Host, rt *goja.Runtime) interface{} {
	return &kafkaModule{host: host, rt: rt, client: host.KafkaClient()}
}

func (m *kafkaModule) Produce(v goja.Value) interface{} {
	args := mapParams(v, m.rt)
	r := &ProduceResult{}
	var err error
	r.Key, r.Value, err = m.client.Produce(args)
	if err != nil {
		log.Errorf("js error: %v in %v", err, m.host.Name())
		r.Error = err.Error()
	}
	return r
}

func mapParams(args goja.Value, rt *goja.Runtime) *common.KafkaProduceArgs {
	opt := &common.KafkaProduceArgs{Partition: -1}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(rt)
		for _, k := range params.Keys() {
			switch k {
			case "cluster":
				cluster := params.Get(k)
				if goja.IsUndefined(cluster) || goja.IsNull(cluster) {
					continue
				}
				opt.Cluster = cluster.String()
			case "topic":
				topic := params.Get(k)
				if goja.IsUndefined(topic) || goja.IsNull(topic) {
					continue
				}
				opt.Topic = topic.String()
			case "partition":
				partition := params.Get(k)
				if goja.IsUndefined(partition) || goja.IsNull(partition) {
					continue
				}
				opt.Partition = int(partition.ToInteger())
			case "key":
				key := params.Get(k)
				if goja.IsUndefined(key) || goja.IsNull(key) {
					continue
				}
				opt.Key = key.Export()
			case "value":
				value := params.Get(k)
				if goja.IsUndefined(value) || goja.IsNull(value) {
					continue
				}
				opt.Value = value.Export()
			case "headers":
				headers := params.Get(k)
				if goja.IsUndefined(headers) || goja.IsNull(headers) {
					continue
				}
				opt.Headers = headers.Export().(map[string]interface{})
			}
		}
	}
	return opt
}
