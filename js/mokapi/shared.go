package mokapi

import (
	"mokapi/engine/common"

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
	return m.store.Get(key)
}

func (m *SharedMemory) Has(key string) bool {
	return m.store.Has(key)
}

func (m *SharedMemory) Set(key string, value any) {
	m.store.Set(key, value)
}

func (m *SharedMemory) Delete(key string) {
	m.store.Delete(key)
}

func (m *SharedMemory) Clear() {
	m.store.Clear()
}

func (m *SharedMemory) Update(key string, fn func(v any) any) any {
	r := m.store.Update(key, func(v any) any {
		return fn(v)
	})

	switch val := r.(type) {
	case map[string]any:
		return m.vm.NewDynamicObject(&SharedObject{m: val, vm: m.vm})
	case []any:
		return m.vm.NewDynamicArray(&SharedArray{array: val, vm: m.vm})
	default:
		return val
	}
}

func (m *SharedMemory) Keys() []string {
	return m.store.Keys()
}

func (m *SharedMemory) Namespace(name string) *SharedMemory {
	s := m.store.Namespace(name)
	return &SharedMemory{store: s}
}

type SharedObject struct {
	m  map[string]any
	vm *goja.Runtime
}

func (v *SharedObject) Get(key string) goja.Value {
	val, ok := v.m[key]
	if !ok {
		return goja.Undefined()
	}
	return toValue(val, v.vm, func(val any) {
		v.m[key] = val
	})
}

func (v *SharedObject) Set(key string, val goja.Value) bool {
	v.m[key] = val.Export()
	return true
}

func (v *SharedObject) Delete(key string) bool {
	if _, ok := v.m[key]; ok {
		delete(v.m, key)
		return true
	}
	return false
}

func (v *SharedObject) Has(key string) bool {
	if _, ok := v.m[key]; ok {
		return true
	}
	return false
}

func (v *SharedObject) Keys() []string {
	var keys []string
	for k := range v.m {
		keys = append(keys, k)
	}
	return keys
}

func (v *SharedObject) Export() any {
	return v.m
}

type SharedArray struct {
	array  []any
	vm     *goja.Runtime
	update func(v any)
}

func (s *SharedArray) Get(idx int) goja.Value {
	if idx < 0 {
		idx += len(s.array)
	}
	if idx >= 0 && idx < len(s.array) {
		return toValue(s.array[idx], s.vm, nil)
	}
	return nil
}

func (s *SharedArray) Set(idx int, val goja.Value) bool {
	if idx < 0 {
		idx += len(s.array)
	}
	if idx < 0 {
		return false
	}
	if idx >= len(s.array) {
		s.expand(idx + 1)
	}
	s.array[idx] = val.Export()
	return true
}

func (s *SharedArray) Len() int {
	return len(s.array)
}

func (s *SharedArray) SetLen(n int) bool {
	if n > len(s.array) {
		s.expand(n)
		return true
	}
	if n < 0 {
		return false
	}
	if n < len(s.array) {
		tail := s.array[n:len(s.array)]
		for j := range tail {
			tail[j] = nil
		}
		s.array = s.array[:n]
		if s.update != nil {
			s.update(s.array)
		}
	}
	return true
}

func (s *SharedArray) expand(newLen int) {
	if newLen > cap(s.array) {
		a := make([]any, newLen)
		copy(a, s.array)
		s.array = a
	} else {
		s.array = s.array[:newLen]
	}
	if s.update != nil {
		s.update(s.array)
	}
}

func (s *SharedArray) Export() any {
	return s.array
}

func toValue(value any, vm *goja.Runtime, update func(v any)) goja.Value {
	switch v := value.(type) {
	case map[string]any:
		return vm.NewDynamicObject(&SharedObject{m: v, vm: vm})
	case []any:
		return vm.NewDynamicArray(&SharedArray{array: v, vm: vm, update: update})
	default:
		return vm.ToValue(value)
	}
}

func Export(v any) any {
	switch val := v.(type) {
	case *SharedObject:
		m := make(map[string]any)
		for k, item := range val.m {
			m[k] = Export(item)
		}
		return m
	case *SharedArray:
		arr := make([]any, len(val.array))
		for i, item := range val.array {
			arr[i] = Export(item)
		}
		return arr
	case goja.Value:
		return Export(val.Export())
	default:
		return v
	}
}
