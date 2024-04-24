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
	if len(result.Records) == 1 {
		deprecatedResult := &kafkaResult{KafkaProduceResult: result}
		deprecatedResult.Key = result.Records[0].Key
		deprecatedResult.Value = result.Records[0].Value
		deprecatedResult.Headers = result.Records[0].Headers
		deprecatedResult.Partition = result.Records[0].Partition
	}
	return result
}

func mapParams(args goja.Value, rt *goja.Runtime) *common.KafkaProduceArgs {
	opt := &common.KafkaProduceArgs{}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(rt)
		var record *common.KafkaRecord
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
				if record == nil {
					record = &common.KafkaRecord{}
				}

				partition := params.Get(k)
				if goja.IsUndefined(partition) || goja.IsNull(partition) {
					continue
				}
				record.Partition = int(partition.ToInteger())
			case "key":
				if record == nil {
					record = &common.KafkaRecord{Partition: -1}
				}

				key := params.Get(k)
				if goja.IsUndefined(key) || goja.IsNull(key) {
					continue
				}
				record.Key = key.Export()
			case "value":
				if record == nil {
					record = &common.KafkaRecord{Partition: -1}
				}

				value := params.Get(k)
				if goja.IsUndefined(value) || goja.IsNull(value) {
					continue
				}
				record.Data = value.Export()
			case "headers":
				if record == nil {
					record = &common.KafkaRecord{Partition: -1}
				}

				headers := params.Get(k)
				if goja.IsUndefined(headers) || goja.IsNull(headers) {
					continue
				}
				record.Headers = headers.Export().(map[string]interface{})
			case "records":
				records := params.Get(k).Export().([]interface{})
				for _, item := range records {
					rec := item.(map[string]interface{})
					r := common.KafkaRecord{Partition: -1}
					if k, ok := rec["key"]; ok {
						r.Key = k
					}
					if v, ok := rec["value"]; ok {
						if b, ok := v.([]byte); ok {
							r.Value = b
						} else if s, ok := v.(string); ok {
							r.Value = []byte(s)
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
					opt.Records = append(opt.Records, r)
				}
			}
		}
		if record != nil {
			opt.Records = append(opt.Records, *record)
		} else if len(opt.Records) == 0 {
			opt.Records = append(opt.Records, common.KafkaRecord{})
		}
	}
	return opt
}
