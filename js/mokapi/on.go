package mokapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"mokapi/js/util"
	"reflect"
)

type onArgs struct {
	tags  map[string]string
	track bool
}

func (m *Module) On(event string, do goja.Value, vArgs goja.Value) {
	eventArgs, err := getOnArgs(m.vm, vArgs)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}

	f := func(args ...interface{}) (bool, error) {
		origin, err := getHashes(args...)
		if err != nil {
			return false, err
		}

		var r goja.Value
		r, err = m.loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
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

		if r != goja.Undefined() {
			return r.ToBoolean(), nil
		}

		if eventArgs.track {
			return true, nil
		}

		newHashes, err := getHashes(args...)
		if err != nil {
			return false, err
		}

		return haveChanges(origin, newHashes), nil
	}

	m.host.On(event, f, eventArgs.tags)
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
				if v.ExportType().Kind() != reflect.Bool {
					return onArgs{}, fmt.Errorf("unexpected type for track: %v", util.JsType(v.Export()))
				}
				result.track = v.ToBoolean()
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
			return nil, fmt.Errorf("unable to marshal arg")
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
