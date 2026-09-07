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
	callback, result := publish(f, vMsg, vArgs, m.m.host.MqttClient(), m.m.vm)
	if callback != nil {
		if err := callback(result); err != nil {
			panic(m.m.vm.ToValue(err.Error()))
		}
	}
	return m
}

func (m *Mqtt) PublishAsync(topic string, vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	p, resolve, reject := m.m.vm.NewPromise()
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				m.m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		f := m.filter
		f.Topic = topic
		callback, result := publish(f, vMsg, vArgs, m.m.host.MqttClient(), m.m.vm)
		m.m.loop.Run(func(vm *goja.Runtime) {
			if callback != nil {
				if err := callback(result); err != nil {
					panic(vm.ToValue(err.Error()))
				}
			}
			_ = resolve(m)
		})
	}()
	return p
}

func (m *MqttTopic) Publish(vMsg goja.Value, vArgs goja.Value) *MqttTopic {
	callback, result := publish(m.filter, vMsg, vArgs, m.m.host.MqttClient(), m.m.vm)
	if callback != nil {
		if err := callback(result); err != nil {
			panic(m.m.vm.ToValue(err.Error()))
		}
	}
	return m
}

func (m *MqttTopic) PublishAsync(vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	p, resolve, reject := m.m.vm.NewPromise()
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				m.m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		callback, result := publish(m.filter, vMsg, vArgs, m.m.host.MqttClient(), m.m.vm)
		m.m.loop.Run(func(vm *goja.Runtime) {
			if callback != nil {
				if err := callback(result); err != nil {
					panic(vm.ToValue(err.Error()))
				}
			}
			_ = resolve(m)
		})
	}()
	return p
}

func publish(filter common.MqttFilter, vMsg goja.Value, vArgs goja.Value, client common.MqttClient, vm *goja.Runtime) (MqttPublishResultCallback, *common.MqttPublishResult) {
	defer func() {
		r := recover()
		if r != nil {
			panic(vm.ToValue(fmt.Sprintf("%v", r)))
		}
	}()

	msg, err := mapPublishMessage(vMsg.ToObject(vm))
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}
	args, err := mapPublishArgs(vArgs, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	file := util.GetScriptFile(vm)
	result, err := client.Publish(&common.MqttPublishArgs{
		Cluster:    filter.Api,
		Topic:      filter.Topic,
		Data:       msg.Data,
		Value:      msg.Value,
		Retain:     msg.Retain,
		Retry:      args.Retry,
		ClientId:   "mokapi-script",
		ScriptFile: file.Info.Key(),
	})
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	return args.result, result
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
				return nil, fmt.Errorf("unexpected type for 'fake': expected function, got %s", util.JsType(jsObj.Get(propName)))
			}
			args.result = func(result *common.MqttPublishResult) error {
				_, err := call(goja.Undefined(), vm.ToValue(result))
				return err
			}
		}
	}

	return args, nil
}
