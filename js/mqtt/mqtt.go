package mqtt

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"time"

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
	defer func() {
		r := recover()
		if r != nil {
			panic(m.rt.ToValue(fmt.Sprintf("%v", r)))
		}
	}()

	args, err := m.mapParams(v)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	client := m.host.MqttClient()

	result, err := client.Publish(args)
	if err != nil {
		log.Errorf("js error: %v in %v", err, m.host.Name())
		panic(m.rt.ToValue(err.Error()))
	}
	return result
}

func (m *Module) PublishAsync(v goja.Value) interface{} {
	p, resolve, reject := m.rt.NewPromise()
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		result := m.Publish(v)
		m.loop.Run(func(vm *goja.Runtime) {
			_ = resolve(result)
		})
	}()
	return p
}

func (m *Module) mapParams(args goja.Value) (*common.MqttPublishArgs, error) {
	file := getFile(m.rt)
	pa := &common.MqttPublishArgs{
		ClientId:   "mokapi-script",
		ScriptFile: file.Info.Key(),
		Retry: common.RetryArgs{
			MaxRetryTime:     3 * time.Minute,
			InitialRetryTime: 500 * time.Millisecond,
			Retries:          10,
			Factor:           2,
		},
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
			case "retain":
				v := params.Get(k)
				pa.Retain = v.ToBoolean()
			case "value":
				value := params.Get(k)
				if goja.IsUndefined(value) || goja.IsNull(value) {
					continue
				}
				pa.Value = value.String()
			case "retry":
				retry := params.Get(k).Export().(map[string]interface{})
				if i, ok := retry["maxRetryTime"]; ok {
					switch v := i.(type) {
					case int64:
						pa.Retry.MaxRetryTime = time.Duration(v) * time.Millisecond
					case string:
						d, err := time.ParseDuration(v)
						if err != nil {
							return nil, fmt.Errorf("parse maxRetryTime failed: %w", err)
						}
						pa.Retry.MaxRetryTime = d
					default:
						return nil, fmt.Errorf("type %T for maxRetryTime not supported", v)
					}

				}
				if i, ok := retry["initialRetryTime"]; ok {
					switch v := i.(type) {
					case int64:
						pa.Retry.InitialRetryTime = time.Duration(v) * time.Millisecond
					case string:
						d, err := time.ParseDuration(v)
						if err != nil {
							return nil, fmt.Errorf("parse initialRetryTime failed: %w", err)
						}
						pa.Retry.InitialRetryTime = d
					default:
						return nil, fmt.Errorf("type %T for initialRetryTime not supported", v)
					}
				}
				if v, ok := retry["retries"]; ok {
					pa.Retry.Retries = int(v.(int64))
				}
				if v, ok := retry["factor"]; ok {
					pa.Retry.Factor = int(v.(int64))
				}
			}
		}
	}
	return pa, nil
}

func getFile(vm *goja.Runtime) *dynamic.Config {
	return vm.Get("mokapi/internal").(*goja.Object).Get("file").Export().(*dynamic.Config)
}
