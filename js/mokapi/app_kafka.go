package mokapi

import (
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"mokapi/js/util"

	"github.com/dop251/goja"
)

type Kafka struct {
	filter common.KafkaFilter
	m      *Module
}

type KafkaTopic struct {
	filter common.KafkaFilter
	m      *Module
}

type KafkaProduceArgs struct {
	Retry  common.RetryArgs
	result KafkaProduceResultCallback
}

type KafkaProduceResultCallback func(result common.KafkaMessageResult) error

func (k *Kafka) Topic(topicName string) *KafkaTopic {
	f := k.filter
	f.Topic = topicName
	return &KafkaTopic{filter: f, m: k.m}
}

func (k *Kafka) Api(name string) *Kafka {
	f := k.filter
	f.Api = name
	return &Kafka{filter: f, m: k.m}
}

func (k *Kafka) Message(do goja.Value, vArgs goja.Value) *Kafka {
	args, err := getOnArgs(k.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, k.m.vm, k.m.loop)

	k.m.host.OnKafka(k.filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return k
}

func (k *KafkaTopic) Message(do goja.Value, vArgs goja.Value) *KafkaTopic {
	args, err := getOnArgs(k.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, k.m.vm, k.m.loop)

	k.m.host.OnKafka(k.filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return k
}

func (k *Kafka) Produce(topic string, vMsg goja.Value, vArgs goja.Value) *Kafka {
	f := k.filter
	f.Topic = topic
	produce(k.m, f, vMsg, vArgs)
	return k
}

func (k *Kafka) ProduceAsync(topic string, vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	f := k.filter
	f.Topic = topic
	return produceAsync(k.m, f, vMsg, vArgs, k)
}

func (k *KafkaTopic) Produce(vMsg goja.Value, vArgs goja.Value) *KafkaTopic {
	produce(k.m, k.filter, vMsg, vArgs)
	return k
}

func (k *KafkaTopic) ProduceAsync(vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	return produceAsync(k.m, k.filter, vMsg, vArgs, k)
}

func produce(m *Module, filter common.KafkaFilter, vMsg goja.Value, vArgs goja.Value) {
	args, callback := prepareKafkaProduce(m.vm, filter, vMsg, vArgs)
	result, err := m.host.KafkaClient().Produce(args)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	if callback != nil {
		for _, r := range result.Messages {
			if err := callback(r); err != nil {
				panic(m.vm.ToValue(err.Error()))
			}
		}
	}
}

// produceAsync resolves the promise with self, like the synchronous variant returns it.
func produceAsync[T any](m *Module, filter common.KafkaFilter, vMsg goja.Value, vArgs goja.Value, self T) *goja.Promise {
	args, callback := prepareKafkaProduce(m.vm, filter, vMsg, vArgs)
	client := m.host.KafkaClient()
	return util.RunAsync(m.vm, m.loop,
		func() (*common.KafkaProduceResult, error) { return client.Produce(args) },
		func(result *common.KafkaProduceResult) (any, error) {
			if callback != nil {
				for _, r := range result.Messages {
					if err := callback(r); err != nil {
						return nil, err
					}
				}
			}
			return self, nil
		})
}

// prepareKafkaProduce converts messages and arguments to plain Go values. It has to run on
// the event loop because it reads from the goja runtime. Invalid input throws.
func prepareKafkaProduce(vm *goja.Runtime, filter common.KafkaFilter, vMsg goja.Value, vArgs goja.Value) (*common.KafkaProduceArgs, KafkaProduceResultCallback) {
	defer util.ThrowPanic(vm)

	msg, err := mapProduceMessages(vMsg, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}
	args, err := mapProduceArgs(vArgs, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	return &common.KafkaProduceArgs{
		Cluster:    filter.Api,
		Topic:      filter.Topic,
		Messages:   msg,
		Retry:      args.Retry,
		ClientId:   "mokapi-script",
		ScriptFile: util.GetScriptFile(vm).Info.Key(),
	}, args.result
}

func mapProduceMessages(arg goja.Value, vm *goja.Runtime) ([]common.KafkaMessage, error) {
	var result []common.KafkaMessage
	if arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
		jsObj := arg.ToObject(vm)
		if jsObj.ClassName() == "Array" {
			lengthKey := jsObj.Get("length")
			length := lengthKey.ToInteger()
			for i := int64(0); i < length; i++ {
				item := jsObj.Get(fmt.Sprintf("%d", i))
				msg, err := mapProduceMessage(item.ToObject(vm))
				if err != nil {
					return nil, err
				}
				result = append(result, msg)
			}
		} else {
			msg, err := mapProduceMessage(jsObj)
			if err != nil {
				return nil, err
			}
			result = append(result, msg)
		}
	}
	return result, nil
}

func mapProduceMessage(vMsg *goja.Object) (common.KafkaMessage, error) {
	msg := common.KafkaMessage{Partition: -1}

	if vMsg == nil || goja.IsUndefined(vMsg) || goja.IsNull(vMsg) {
		return msg, nil
	}

	for _, propName := range vMsg.Keys() {
		switch propName {
		case "key":
			msg.Key = vMsg.Get(propName).Export()
		case "data":
			msg.Data = common.Optional[any]{Value: vMsg.Get(propName).Export(), IsSet: true}
		case "value":
			v := vMsg.Get(propName).Export()
			switch val := v.(type) {
			case goja.ArrayBuffer:
				msg.Value = common.Optional[[]byte]{Value: val.Bytes(), IsSet: true}
			case string, int64, float64, bool:
				msg.Value = common.Optional[[]byte]{Value: []byte(fmt.Sprintf("%v", val)), IsSet: true}
			case nil:
				msg.Value = common.Optional[[]byte]{Value: nil, IsSet: true}
			default:
				b, err := json.Marshal(val)
				if err != nil {
					return msg, fmt.Errorf("kafka produce: value cannot be serialized: %v", err)
				}
				msg.Value = common.Optional[[]byte]{Value: b, IsSet: true}

			}
		case "headers":
			headers := vMsg.Get(propName).Export().(map[string]any)
			msg.Headers = map[string]string{}
			for key, val := range headers {
				switch v := val.(type) {
				case string, int64, float64, bool:
					msg.Headers[key] = fmt.Sprintf("%v", v)
				case nil:
					msg.Headers[key] = ""
				default:
					b, err := json.Marshal(val)
					if err != nil {
						return msg, fmt.Errorf("kafka produce: value cannot be serialized: %v", err)
					}
					msg.Headers[key] = string(b)
				}
			}
		}
	}
	return msg, nil
}

func mapProduceArgs(v goja.Value, vm *goja.Runtime) (*KafkaProduceArgs, error) {
	args := &KafkaProduceArgs{
		Retry: util.DefaultRetryArgs(),
	}

	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return args, nil
	}

	jsObj := v.ToObject(vm)
	for _, propName := range jsObj.Keys() {
		switch propName {
		case "retry":
			retry, ok := jsObj.Get(propName).Export().(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid property type '%s'", propName)
			}
			var err error
			args.Retry, err = util.ConvertToRetryArgs(retry)
			if err != nil {
				return args, err
			}
		case "result":
			call, ok := goja.AssertFunction(jsObj.Get(propName))
			if !ok {
				return nil, fmt.Errorf("unexpected type for 'produce': expected function, got %s", util.JsType(jsObj.Get(propName)))
			}
			args.result = func(result common.KafkaMessageResult) error {
				_, err := call(goja.Undefined(), vm.ToValue(result))
				return err
			}
		}
	}

	return args, nil
}
