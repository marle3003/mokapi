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
	Cluster   string `json:"cluster"`
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

func newKafka(host common.Host, rt *goja.Runtime) interface{} {
	return &kafkaModule{host: host, rt: rt, client: host.KafkaClient()}
}

func (m *kafkaModule) Produce(v goja.Value) interface{} {
	args := mapParams(v, m.rt)
	result, err := m.client.Produce(args)
	if err != nil {
		log.Errorf("js error: %v in %v", err, m.host.Name())
		panic(m.rt.ToValue(err.Error()))
	}
	return ProduceResult{
		Cluster:   result.Cluster,
		Topic:     result.Topic,
		Partition: result.Partition,
		Offset:    result.Offset,
		Key:       result.Key,
		Value:     result.Value,
	}
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
