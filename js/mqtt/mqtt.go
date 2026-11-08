package mqtt

import (
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"mokapi/js/util"

	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
)

type Module struct {
	host common.Host
	rt   *goja.Runtime
	loop *eventloop.EventLoop
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
	_ = obj.Set("publish", f.Publish)
	_ = obj.Set("publishAsync", f.PublishAsync)
}

func (m *Module) Publish(v goja.Value) interface{} {
	defer util.ThrowPanic(m.rt)

	args := m.preparePublish(v)
	result, err := m.execute(args)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return result
}

func (m *Module) PublishAsync(v goja.Value) interface{} {
	args := m.preparePublish(v)
	return util.RunAsync(m.rt, m.loop, func() (*common.MqttPublishResult, error) {
		return m.execute(args)
	}, nil)
}

// preparePublish converts the JavaScript arguments to plain Go values. It has to run on the
// event loop because it reads from the goja runtime. Invalid input throws.
func (m *Module) preparePublish(v goja.Value) *common.MqttPublishArgs {
	defer util.ThrowPanic(m.rt)

	args, err := m.mapParams(v)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return args
}

// execute only calls the MQTT client and must not touch the goja runtime.
func (m *Module) execute(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
	result, err := m.host.MqttClient().Publish(args)
	if err != nil {
		log.Errorf("js error: %v in %v", err, m.host.Name())
		return nil, err
	}
	return result, nil
}

func (m *Module) mapParams(args goja.Value) (*common.MqttPublishArgs, error) {
	pa := &common.MqttPublishArgs{
		ClientId:   "mokapi-script",
		ScriptFile: util.GetScriptFile(m.rt).Info.Key(),
		Retry:      util.DefaultRetryArgs(),
	}

	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "cluster":
				cluster := params.Get(k)
				if goja.IsUndefined(cluster) || goja.IsNull(cluster) {
					continue
				}
				pa.Cluster = cluster.String()
			case "topic":
				topic := params.Get(k)
				if goja.IsUndefined(topic) || goja.IsNull(topic) {
					continue
				}
				pa.Topic = topic.String()
			case "messageId":
				var err error
				pa.MessageId, err = util.MessageId(params.Get(k).Export())
				if err != nil {
					return nil, err
				}
			case "retain":
				v := params.Get(k)
				pa.Retain = v.ToBoolean()
			case "value":
				value := params.Get(k)
				if goja.IsUndefined(value) {
					continue
				}
				if goja.IsNull(value) {
					// an empty payload, not nil: nil asks for a generated payload
					pa.Value = []byte{}
					continue
				}
				pa.Value = []byte(value.String())
			case "retry":
				retry := params.Get(k).Export().(map[string]interface{})
				var err error
				pa.Retry, err = util.ConvertToRetryArgs(retry)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return pa, nil
}
