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
	callback, result := produce(f, vMsg, vArgs, k.m.host.KafkaClient(), k.m.vm)
	if callback != nil {
		for _, r := range result.Messages {
			if err := callback(r); err != nil {
				panic(k.m.vm.ToValue(err.Error()))
			}
		}
	}
	return k
}

func (k *Kafka) ProduceAsync(topic string, vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	p, resolve, reject := k.m.vm.NewPromise()
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				k.m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		f := k.filter
		f.Topic = topic
		callback, result := produce(f, vMsg, vArgs, k.m.host.KafkaClient(), k.m.vm)
		k.m.loop.Run(func(vm *goja.Runtime) {
			if callback != nil {
				for _, r := range result.Messages {
					if err := callback(r); err != nil {
						panic(vm.ToValue(err.Error()))
					}
				}
			}
			_ = resolve(k)
		})
	}()
	return p
}

func (k *KafkaTopic) Produce(vMsg goja.Value, vArgs goja.Value) *KafkaTopic {
	callback, result := produce(k.filter, vMsg, vArgs, k.m.host.KafkaClient(), k.m.vm)
	if callback != nil {
		for _, r := range result.Messages {
			if err := callback(r); err != nil {
				panic(k.m.vm.ToValue(err.Error()))
			}
		}
	}
	return k
}

func (k *KafkaTopic) ProduceAsync(vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	p, resolve, reject := k.m.vm.NewPromise()
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				k.m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		callback, result := produce(k.filter, vMsg, vArgs, k.m.host.KafkaClient(), k.m.vm)
		k.m.loop.Run(func(vm *goja.Runtime) {
			if callback != nil {
				for _, r := range result.Messages {
					if err := callback(r); err != nil {
						panic(vm.ToValue(err.Error()))
					}
				}
			}
			_ = resolve(k)
		})
	}()
	return p
}

func produce(filter common.KafkaFilter, vMsg goja.Value, vArgs goja.Value, client common.KafkaClient, vm *goja.Runtime) (KafkaProduceResultCallback, *common.KafkaProduceResult) {
	defer func() {
		r := recover()
		if r != nil {
			panic(vm.ToValue(fmt.Sprintf("%v", r)))
		}
	}()

	msg, err := mapProduceMessages(vMsg, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}
	args, err := mapProduceArgs(vArgs, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	file := util.GetScriptFile(vm)
	result, err := client.Produce(&common.KafkaProduceArgs{
		Cluster:    filter.Api,
		Topic:      filter.Topic,
		Messages:   msg,
		Retry:      args.Retry,
		ClientId:   "mokapi-script",
		ScriptFile: file.Info.Key(),
	})
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	return args.result, result
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
			msg.Data = vMsg.Get(propName).Export()
		case "value":
			v := vMsg.Get(propName).Export()
			switch val := v.(type) {
			case goja.ArrayBuffer:
				msg.Value = val.Bytes()
			case string, int64, float64, bool:
				msg.Value = []byte(fmt.Sprintf("%v", val))
			case nil:
				msg.Value = nil
			default:
				b, err := json.Marshal(val)
				if err != nil {
					return msg, fmt.Errorf("kafka produce: value cannot be serialized: %v", err)
				}
				msg.Value = b

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
				return nil, fmt.Errorf("unexpected type for 'fake': expected function, got %s", util.JsType(jsObj.Get(propName)))
			}
			args.result = func(result common.KafkaMessageResult) error {
				_, err := call(goja.Undefined(), vm.ToValue(result))
				return err
			}
		}
	}

	return args, nil
}
