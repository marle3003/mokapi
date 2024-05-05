package kafka

import (
	"fmt"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"time"
)

type Module struct {
	host common.Host
	rt   *goja.Runtime
	loop *eventloop.EventLoop
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

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	loop := o.Get("loop").Export().(*eventloop.EventLoop)
	f := &Module{
		rt:   vm,
		host: host,
		loop: loop,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("produce", f.Produce)
	obj.Set("produceAsync", f.ProduceAsync)
}

func (m *Module) Produce(v goja.Value) interface{} {
	args, err := m.mapParams(v)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	client := m.host.KafkaClient()

	result, err := client.Produce(args)
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

func (m *Module) ProduceAsync(v goja.Value) interface{} {
	p, resolve, _ := m.rt.NewPromise()
	go func() {
		result := m.Produce(v)
		m.loop.Run(func(vm *goja.Runtime) {
			resolve(result)
		})
	}()
	return p
}

func (m *Module) mapParams(args goja.Value) (*common.KafkaProduceArgs, error) {
	opt := &common.KafkaProduceArgs{Retry: common.KafkaProduceRetry{
		MaxRetryTime:     30000 * time.Millisecond,
		InitialRetryTime: 200 * time.Millisecond,
		Retries:          5,
		Factor:           4,
	}}

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
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
				m.warnDeprecatedAttribute("partition")

				partition := params.Get(k)
				if goja.IsUndefined(partition) || goja.IsNull(partition) {
					continue
				}
				message.Partition = int(partition.ToInteger())
			case "key":
				if message == nil {
					message = &common.KafkaMessage{Partition: -1}
				}
				m.warnDeprecatedAttribute("key")

				key := params.Get(k)
				if goja.IsUndefined(key) || goja.IsNull(key) {
					continue
				}
				message.Key = key.Export()
			case "value":
				if message == nil {
					message = &common.KafkaMessage{Partition: -1}
				}
				m.warnDeprecatedAttribute("value")

				value := params.Get(k)
				if goja.IsUndefined(value) || goja.IsNull(value) {
					continue
				}
				message.Data = value.Export()
			case "headers":
				if message == nil {
					message = &common.KafkaMessage{Partition: -1}
				}
				m.warnDeprecatedAttribute("headers")

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
			case "retry":
				retry := params.Get(k).Export().(map[string]interface{})
				if i, ok := retry["maxRetryTime"]; ok {
					switch v := i.(type) {
					case int64:
						opt.Retry.MaxRetryTime = time.Duration(v) * time.Millisecond
					case string:
						d, err := time.ParseDuration(v)
						if err != nil {
							return nil, fmt.Errorf("parse maxRetryTime failed: %w", err)
						}
						opt.Retry.MaxRetryTime = d
					default:
						return nil, fmt.Errorf("type %T for maxRetryTime not supported", v)
					}

				}
				if i, ok := retry["initialRetryTime"]; ok {
					switch v := i.(type) {
					case int64:
						opt.Retry.InitialRetryTime = time.Duration(v) * time.Millisecond
					case string:
						d, err := time.ParseDuration(v)
						if err != nil {
							return nil, fmt.Errorf("parse initialRetryTime failed: %w", err)
						}
						opt.Retry.InitialRetryTime = d
					default:
						return nil, fmt.Errorf("type %T for initialRetryTime not supported", v)
					}
				}
				if v, ok := retry["retries"]; ok {
					opt.Retry.Retries = int(v.(int64))
				}
				if v, ok := retry["factor"]; ok {
					opt.Retry.Factor = int(v.(int64))
				}
			}
		}
		if message != nil {
			opt.Messages = append(opt.Messages, *message)
		} else if len(opt.Messages) == 0 {
			opt.Messages = append(opt.Messages, common.KafkaMessage{})
		}
	}
	return opt, nil
}

func (m *Module) warnDeprecatedAttribute(name string) {
	m.host.Warn(fmt.Sprintf("DEPRECATED: '%v' should not be used anymore: check https://mokapi.io/docs/javascript-api/mokapi-kafka/produceargs for more info in %v", name, m.host.Name()))
}
