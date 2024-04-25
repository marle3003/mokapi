package js

import (
	"fmt"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
)

type kafkaModule struct {
	host   common.Host
	rt     *goja.Runtime
	client common.KafkaClient
}

type kafkaResult struct {
	*common.KafkaProduceResult

	// deprecated fields
	Key       interface{}
	Value     interface{}
	Offset    int64
	Headers   map[string]string
	Partition int
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
	if len(result.Messages) == 1 {
		deprecatedResult := &kafkaResult{KafkaProduceResult: result}
		deprecatedResult.Key = result.Messages[0].Key
		deprecatedResult.Value = result.Messages[0].Value
		deprecatedResult.Headers = result.Messages[0].Headers
		deprecatedResult.Partition = result.Messages[0].Partition
	}
	return result
}

func mapParams(args goja.Value, rt *goja.Runtime) *common.KafkaProduceArgs {
	opt := &common.KafkaProduceArgs{}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(rt)
		var message *common.KafkaMessage
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
				if message == nil {
					message = &common.KafkaMessage{}
				}

				partition := params.Get(k)
				if goja.IsUndefined(partition) || goja.IsNull(partition) {
					continue
				}
				message.Partition = int(partition.ToInteger())
			case "key":
				if message == nil {
					message = &common.KafkaMessage{Partition: -1}
				}

				key := params.Get(k)
				if goja.IsUndefined(key) || goja.IsNull(key) {
					continue
				}
				message.Key = key.Export()
			case "value":
				if message == nil {
					message = &common.KafkaMessage{Partition: -1}
				}

				value := params.Get(k)
				if goja.IsUndefined(value) || goja.IsNull(value) {
					continue
				}
				message.Data = value.Export()
			case "headers":
				if message == nil {
					message = &common.KafkaMessage{Partition: -1}
				}

				headers := params.Get(k)
				if goja.IsUndefined(headers) || goja.IsNull(headers) {
					continue
				}
				message.Headers = headers.Export().(map[string]interface{})
			case "messages":
				records := params.Get(k).Export().([]interface{})
				for _, item := range records {
					rec := item.(map[string]interface{})
					r := common.KafkaMessage{Partition: -1}
					if k, ok := rec["key"]; ok {
						r.Key = k
					}
					if v, ok := rec["value"]; ok {
						if b, ok := v.([]byte); ok {
							r.Value = b
						} else if s, ok := v.(string); ok {
							r.Value = []byte(s)
						} else {
							r.Value = []byte(fmt.Sprintf("%v", v))
						}
					}
					if d, ok := rec["data"]; ok {
						r.Data = d
					}
					if h, ok := rec["headers"]; ok {
						if header, ok := h.(map[string]interface{}); ok {
							r.Headers = header
						}
					}
					if p, ok := rec["partition"]; ok {
						if i, ok := p.(int64); ok {
							r.Partition = int(i)
						}
					}
					opt.Messages = append(opt.Messages, r)
				}
			}
		}
		if message != nil {
			opt.Messages = append(opt.Messages, *message)
		} else if len(opt.Messages) == 0 {
			opt.Messages = append(opt.Messages, common.KafkaMessage{})
		}
	}
	return opt
}
