package mokapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"mokapi/js/util"
	"net/http"
	"reflect"

	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
)

type onArgs struct {
	tags       map[string]string
	track      func(args ...goja.Value) (bool, error)
	isTrackSet bool
	priority   int
}

func (m *Module) On(event string, do goja.Value, vArgs goja.Value) {
	eventArgs, err := getOnArgs(m.vm, vArgs)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}

	f := getHandler(do, eventArgs, m.vm, m.loop)

	switch event {
	case "http":
		m.host.OnHttp(common.HttpFilter{}, f, common.EventArgs{Tags: eventArgs.tags, Priority: eventArgs.priority})
	case "kafka":
		m.host.OnKafka(common.KafkaFilter{}, f, common.EventArgs{Tags: eventArgs.tags, Priority: eventArgs.priority})
	case "mqtt":
		m.host.OnMqtt(common.MqttFilter{}, f, common.EventArgs{Tags: eventArgs.tags, Priority: eventArgs.priority})
	case "websocket":
		m.host.OnWebsocket(common.WebsocketFilter{}, f, common.EventArgs{Tags: eventArgs.tags, Priority: eventArgs.priority})
	case "smtp":
		m.host.OnMail(common.MailFilter{}, f, common.EventArgs{Tags: eventArgs.tags, Priority: eventArgs.priority})
	case "ldap":
		m.host.OnLdap(common.LdapFilter{}, f, common.EventArgs{Tags: eventArgs.tags, Priority: eventArgs.priority})
	default:
		log.Error(fmt.Errorf("unknown event: %s", event))
	}
}

func getHandler(do goja.Value, args onArgs, vm *goja.Runtime, loop *eventloop.EventLoop) common.EventHandler {
	return func(ctx *common.EventContext) (bool, error) {
		origin, err := getHashes(ctx.Args...)
		if err != nil {
			return false, err
		}

		var r goja.Value
		var params []goja.Value
		r, err = loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
			for _, v := range ctx.Args {
				params = append(params, ArgToJs(v, vm))
			}

			call, _ := goja.AssertFunction(do)
			v, err := call(goja.Undefined(), params...)
			if err != nil {
				return nil, err
			}
			return v, nil
		}, &eventloop.JobContext{EventLogger: ctx.EventLogger})

		if err != nil {
			return false, err
		}

		if r != goja.Undefined() {
			return r.ToBoolean(), nil
		}

		if args.isTrackSet {
			return args.track(params...)
		}

		newHashes, err := getHashes(ctx.Args...)
		if err != nil {
			return false, err
		}

		return haveChanges(origin, newHashes), nil
	}
}

func getOnArgs(vm *goja.Runtime, args goja.Value) (onArgs, error) {
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		result := onArgs{tags: map[string]string{}}

		if args.ExportType().Kind() != reflect.Map {
			return onArgs{}, fmt.Errorf("unexpected type for args: %v", util.JsType(args.Export()))
		}
		params := args.ToObject(vm)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				if tagsV.ExportType().Kind() != reflect.Map {
					return onArgs{}, fmt.Errorf("unexpected type for tags: %v", util.JsType(tagsV.Export()))
				}
				tagsO := tagsV.ToObject(vm)
				for _, key := range tagsO.Keys() {
					result.tags[key] = tagsO.Get(key).String()
				}
			case "track":
				v := params.Get(k)
				if goja.IsUndefined(v) || goja.IsNull(v) {
					continue
				}
				if v.ExportType().Kind() == reflect.Bool {
					result.isTrackSet = true
					result.track = func(args ...goja.Value) (bool, error) {
						return v.ToBoolean(), nil
					}
				} else if f, ok := goja.AssertFunction(v); ok {
					result.isTrackSet = true
					result.track = func(args ...goja.Value) (bool, error) {
						r, err := f(goja.Undefined(), args...)
						if err != nil {
							return true, fmt.Errorf("failed to call track function: %v", err)
						}
						if r.ExportType().Kind() == reflect.Bool {
							return r.ToBoolean(), nil
						}
						return true, fmt.Errorf("unexpected return type for track: %v", util.JsType(r.Export()))
					}
				} else {
					return onArgs{}, fmt.Errorf("unexpected type for track: %v", util.JsType(v.Export()))
				}
			case "priority":
				v := params.Get(k)
				if goja.IsUndefined(v) || goja.IsNull(v) {
					continue
				}
				if v.ExportType().Kind() != reflect.Int64 {
					return onArgs{}, fmt.Errorf("unexpected type for priority: %v", util.JsType(v.Export()))
				}
				result.priority = int(v.ToInteger())
			}
		}
		return result, nil
	}
	return onArgs{}, nil
}

func getHashes(args ...any) ([][]byte, error) {
	var result [][]byte
	for _, arg := range args {
		b, err := json.Marshal(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arg: %v", err)
		}
		result = append(result, b)
	}
	return result, nil
}

func haveChanges(origin [][]byte, new [][]byte) bool {
	for i, o := range origin {
		n := new[i]
		if !bytes.Equal(o, n) {
			return true
		}
	}
	return false
}

func ArgToJs(arg any, vm *goja.Runtime) goja.Value {
	switch v := (arg).(type) {
	case *common.HttpEventResponse:
		return vm.NewDynamicObject(&Proxy{
			target: reflect.ValueOf(v),
			vm:     vm,
			ToJSValue: func(vm *goja.Runtime, key string, val any) goja.Value {
				p := NewProxy(val, vm)

				switch key {
				case "headers":
					p.KeyNormalizer = http.CanonicalHeaderKey
				case "rebuild":
					return rebuild(vm, v)
				}

				switch val.(type) {
				case string, int, bool:
					return p.vm.ToValue(val)
				}

				return vm.NewDynamicObject(p)
			},
		})
	default:
		return vm.ToValue(v)
	}
}

func rebuild(vm *goja.Runtime, res *common.HttpEventResponse) goja.Value {
	if res.Rebuild == nil {
		return vm.ToValue(func() {})
	}
	return vm.ToValue(func(statusCode goja.Value, contentType goja.Value) {
		s := int64(0)
		c := ""
		if statusCode != nil {
			if statusCode.ExportType().Kind() != reflect.Int64 {
				panic(fmt.Sprintf("response.rebuild failed: statusCode must be a number: got %v", util.JsType(statusCode.Export())))
			} else {
				s = statusCode.ToInteger()
			}
		}
		if contentType != nil {
			if contentType.ExportType().Kind() != reflect.String {
				panic(fmt.Sprintf("response.rebuild failed: contentType must be a string: got %v", util.JsType(contentType.Export())))
			} else {
				c = contentType.String()
			}
		}
		res.Rebuild(int(s), c)
	})
}
