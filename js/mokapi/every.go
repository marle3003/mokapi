package mokapi

import (
	"github.com/dop251/goja"
)

func (m *Module) Every(every string, do goja.Value, args goja.Value) (int, error) {
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

	return m.host.Every(every, f, options)
}
