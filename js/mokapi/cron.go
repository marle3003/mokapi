package mokapi

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/js/util"
	"reflect"
)

func (m *Module) Cron(expr string, do goja.Value, args goja.Value) (int, error) {
	options, err := getJobOptions(m.vm, args)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}

	f := func() {
		_, err := m.loop.RunAsync(func(vm *goja.Runtime) (goja.Value, error) {
			call, _ := goja.AssertFunction(do)
			v, err := call(goja.Undefined())
			if err != nil {
				return nil, err
			}
			return v, nil
		})
		if err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
	}

	return m.host.Cron(expr, f, options)
}

func getJobOptions(vm *goja.Runtime, opt goja.Value) (common.JobOptions, error) {
	options := common.NewJobOptions()

	if opt != nil && !goja.IsUndefined(opt) && !goja.IsNull(opt) {
		if opt.ExportType().Kind() != reflect.Map {
			return options, fmt.Errorf("unexpected type for args: %v", util.JsType(opt.Export()))
		}

		params := opt.ToObject(vm)
		for _, k := range params.Keys() {
			switch k {
			case "tags":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				if tagsV.ExportType().Kind() != reflect.Map {
					return options, fmt.Errorf("unexpected type for tags: %v", util.JsType(tagsV.Export()))
				}

				tags := tagsV.ToObject(vm)
				for _, key := range tags.Keys() {
					options.Tags[key] = tags.Get(key).String()
				}
			case "times":
				times := params.Get(k)
				if times.ExportType().Kind() != reflect.Int64 {
					return options, fmt.Errorf("unexpected type for times: %v", util.JsType(times.Export()))
				}
				options.Times = int(times.ToInteger())
			case "skipImmediateFirstRun":
				skip := params.Get(k)
				if skip.ExportType().Kind() != reflect.Bool {
					return options, fmt.Errorf("unexpected type for skipImmediateFirstRun: %v", util.JsType(skip.Export()))
				}
				options.SkipImmediateFirstRun = skip.ToBoolean()
			}
		}
	}

	return options, nil
}
