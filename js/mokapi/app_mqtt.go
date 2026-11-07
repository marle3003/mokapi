package mokapi

import (
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"mokapi/js/util"

	"github.com/dop251/goja"
)

type Mqtt struct {
	filter common.MqttFilter
	m      *Module
}

type MqttTopic struct {
	filter common.MqttFilter
	m      *Module
}

type MqttPublishArgs struct {
	Retry  common.RetryArgs
	result MqttPublishResultCallback
}

type MqttPublishResultCallback func(result *common.MqttPublishResult) error

func (m *Mqtt) Topic(topicName string) *MqttTopic {
	f := m.filter
	f.Topic = topicName
	return &MqttTopic{filter: f, m: m.m}
}

func (m *Mqtt) Api(name string) *Mqtt {
	f := m.filter
	f.Api = name
	return &Mqtt{filter: f, m: m.m}
}

func (m *Mqtt) Message(do goja.Value, vArgs goja.Value) *Mqtt {
	args, err := getOnArgs(m.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, m.m.vm, m.m.loop)

	m.m.host.OnMqtt(m.filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return m
}

func (m *MqttTopic) Message(do goja.Value, vArgs goja.Value) *MqttTopic {
	args, err := getOnArgs(m.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, m.m.vm, m.m.loop)

	m.m.host.OnMqtt(m.filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return m
}

func (m *Mqtt) Publish(topic string, vMsg goja.Value, vArgs goja.Value) *Mqtt {
	f := m.filter
	f.Topic = topic
	publish(m.m, f, vMsg, vArgs)
	return m
}

func (m *Mqtt) PublishAsync(topic string, vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	f := m.filter
	f.Topic = topic
	return publishAsync(m.m, f, vMsg, vArgs, m)
}

func (m *MqttTopic) Publish(vMsg goja.Value, vArgs goja.Value) *MqttTopic {
	publish(m.m, m.filter, vMsg, vArgs)
	return m
}

func (m *MqttTopic) PublishAsync(vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	return publishAsync(m.m, m.filter, vMsg, vArgs, m)
}

func publish(m *Module, filter common.MqttFilter, vMsg goja.Value, vArgs goja.Value) {
	args, callback := prepareMqttPublish(m.vm, filter, vMsg, vArgs)
	result, err := m.host.MqttClient().Publish(args)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	if callback != nil {
		if err := callback(result); err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
	}
}

// publishAsync resolves the promise with self, like the synchronous variant returns it.
func publishAsync[T any](m *Module, filter common.MqttFilter, vMsg goja.Value, vArgs goja.Value, self T) *goja.Promise {
	args, callback := prepareMqttPublish(m.vm, filter, vMsg, vArgs)
	client := m.host.MqttClient()
	return util.RunAsync(m.vm, m.loop,
		func() (*common.MqttPublishResult, error) { return client.Publish(args) },
		func(result *common.MqttPublishResult) (any, error) {
			if callback != nil {
				if err := callback(result); err != nil {
					return nil, err
				}
			}
			return self, nil
		})
}

// prepareMqttPublish converts message and arguments to plain Go values. It has to run on
// the event loop because it reads from the goja runtime. Invalid input throws.
func prepareMqttPublish(vm *goja.Runtime, filter common.MqttFilter, vMsg goja.Value, vArgs goja.Value) (*common.MqttPublishArgs, MqttPublishResultCallback) {
	defer util.ThrowPanic(vm)

	msg, err := mapPublishMessage(vMsg.ToObject(vm))
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}
	args, err := mapPublishArgs(vArgs, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	return &common.MqttPublishArgs{
		Cluster:    filter.Api,
		Topic:      filter.Topic,
		Data:       msg.Data,
		Value:      msg.Value,
		Retain:     msg.Retain,
		Retry:      args.Retry,
		ClientId:   "mokapi-script",
		ScriptFile: util.GetScriptFile(vm).Info.Key(),
	}, args.result
}

func mapPublishMessage(vMsg *goja.Object) (*common.MqttPublishArgs, error) {
	msg := &common.MqttPublishArgs{
		Retry: util.DefaultRetryArgs(),
	}

	if vMsg == nil || goja.IsUndefined(vMsg) || goja.IsNull(vMsg) {
		return msg, nil
	}

	for _, propName := range vMsg.Keys() {
		switch propName {
		case "data":
			msg.Data = vMsg.Get(propName).Export()
		case "value":
			v := vMsg.Get(propName).Export()
			switch val := v.(type) {
			case goja.ArrayBuffer:
				msg.Value = val.Bytes()
			case string:
				msg.Value = []byte(val)
			case int64, float64, bool:
				msg.Value = []byte(fmt.Sprintf("%v", val))
			case nil:
				msg.Value = nil
			default:
				b, err := json.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf("mqtt publish: value cannot be serialized: %v", err)
				}
				msg.Value = b

			}
		case "retain":
			v, ok := vMsg.Get(propName).Export().(bool)
			if !ok {
				return nil, fmt.Errorf("unexpected type for 'retain': expected boolean, got %s", util.JsType(vMsg.Get(propName)))
			}
			msg.Retain = v
		}
	}
	return msg, nil
}

func mapPublishArgs(v goja.Value, vm *goja.Runtime) (*MqttPublishArgs, error) {
	args := &MqttPublishArgs{
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
				return nil, err
			}
		case "result":
			call, ok := goja.AssertFunction(jsObj.Get(propName))
			if !ok {
				return nil, fmt.Errorf("unexpected type for 'publish': expected function, got %s", util.JsType(jsObj.Get(propName)))
			}
			args.result = func(result *common.MqttPublishResult) error {
				_, err := call(goja.Undefined(), vm.ToValue(result))
				return err
			}
		}
	}

	return args, nil
}
