package mokapi

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/js/util"
	"reflect"
)

type onArgs struct {
	tags map[string]string
}

func (m *Module) On(event string, do goja.Value, vArgs goja.Value) {
	args, err := getOnArgs(m.vm, vArgs)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}

	f := func(args ...interface{}) (bool, error) {
		m.host.Lock()
		defer m.host.Unlock()

		r, err := m.loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
			call, _ := goja.AssertFunction(do)
			var params []goja.Value
			for _, v := range args {
				params = append(params, vm.ToValue(v))
			}
			v, err := call(goja.Undefined(), params...)
			if err != nil {
				return nil, err
			}
			return v, nil
		})

		if err != nil {
			return false, err
		}

		return r.ToBoolean(), nil
	}

	m.host.On(event, f, args.tags)
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
			}
		}
		return result, nil
	}
	return onArgs{}, nil
}
