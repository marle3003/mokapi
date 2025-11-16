package mokapi

import (
	"fmt"
	"mokapi/engine/common"
	"reflect"
	"slices"

	"github.com/dop251/goja"
)

type SharedMemory struct {
	store common.Store
	vm    *goja.Runtime
}

func NewSharedMemory(store common.Store, vm *goja.Runtime) *SharedMemory {
	return &SharedMemory{store: store, vm: vm}
}

func (m *SharedMemory) Get(key string) any {
	v := m.store.Get(key)
	if v == nil {
		return nil
	}
	uv := v.(*SharedValue)
	return uv.Use(m.vm).ToValue()
}

func (m *SharedMemory) Has(key string) bool {
	return m.store.Has(key)
}

func (m *SharedMemory) Set(key string, value goja.Value) {
	if value == nil {
		m.store.Set(key, nil)
	} else {
		m.store.Set(key, NewSharedValue(value, m.vm))
	}
}

func (m *SharedMemory) Delete(key string) {
	m.store.Delete(key)
}

func (m *SharedMemory) Clear() {
	m.store.Clear()
}

func (m *SharedMemory) Update(key string, fn goja.Value) any {
	p := m.store.Update(key, func(v any) any {
		var arg goja.Value
		if v != nil {
			arg = v.(*SharedValue).Use(m.vm).ToValue()
		}
		call, ok := goja.AssertFunction(fn)
		if !ok {
			panic(m.vm.ToValue(fmt.Errorf("expected function as parameter")))
		}
		r, err := call(goja.Undefined(), arg)
		if err != nil {
			panic(m.vm.ToValue(err))
		}

		return NewSharedValue(r, m.vm)
	})
	return p.(*SharedValue).ToValue()
}

func (m *SharedMemory) Keys() []string {
	return m.store.Keys()
}

func (m *SharedMemory) Namespace(name string) *SharedMemory {
	s := m.store.Namespace(name)
	return &SharedMemory{store: s, vm: m.vm}
}

func Export(v any) any {
	switch val := v.(type) {
	case *Proxy:
		return val.Export()
	case *SharedValue:
		return val.source.Export()
	case goja.Value:
		return Export(val.Export())
	default:
		return v
	}
}

// SharedValue represents a Go-managed value that can be shared across
// multiple Goja runtimes, while maintaining reference identity.
type SharedValue struct {
	vm     *goja.Runtime
	source goja.Value
}

func NewSharedValue(v goja.Value, vm *goja.Runtime) *SharedValue {
	return &SharedValue{
		source: v,
		vm:     vm,
	}
}

func (p *SharedValue) Use(vm *goja.Runtime) *SharedValue {
	return &SharedValue{source: p.source, vm: vm}
}

func (p *SharedValue) Get(key string) goja.Value {
	switch v := p.source.(type) {
	case *goja.Object:
		f := v.Get(key)
		if _, ok := goja.AssertFunction(f); ok {
			return f
		} else if _, isObject := f.(*goja.Object); isObject {
			return p.vm.NewDynamicObject(NewSharedValue(f, p.vm))
		}
		return f
	}

	return goja.Undefined()
}

func (p *SharedValue) Has(key string) bool {
	switch v := p.source.(type) {
	case *goja.Object:
		return slices.Contains(v.Keys(), key)
	default:
		return false
	}
}

func (p *SharedValue) Set(key string, value goja.Value) bool {
	switch v := p.source.(type) {
	case *goja.Object:
		err := v.Set(key, value)
		if err != nil {
			panic(p.vm.ToValue(err))
		}
		return true
	}
	return false
}

func (p *SharedValue) Delete(key string) bool {
	switch v := p.source.(type) {
	case *goja.Object:
		err := v.Delete(key)
		if err != nil {
			panic(p.vm.ToValue(err))
		}
		return true
	default:
		return false
	}
}

func (p *SharedValue) Keys() []string {
	switch v := p.source.(type) {
	case *goja.Object:
		return v.Keys()
	default:
		return nil
	}
}

func (p *SharedValue) ToValue() goja.Value {
	if p.source == nil {
		return goja.Undefined()
	}
	switch p.source.ExportType().Kind() {
	case reflect.Map, reflect.Slice:
		return p.vm.NewDynamicObject(p)
	default:
		return p.source
	}
}
