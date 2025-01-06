package mokapi

import (
	"github.com/dop251/goja"
)

func (m *Module) Every(every string, do func(), args goja.Value) (int, error) {
	options, err := getJobOptions(m.vm, args)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}

	f := func() {
		m.host.Lock()
		defer m.host.Unlock()
		do()
	}

	return m.host.Every(every, f, options)
}
