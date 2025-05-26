package faker

import (
	"github.com/dop251/goja"
	"strconv"
)

type TrackObject[T any] interface {
	convert(v goja.Value) T
}

type JsArrayAccessor[T any] interface {
	Len() int
	Get(i int) T
	Set(i int, item T)
	Splice(start int, deleteCount int, items []T)
}

type TrackArray[T any] struct {
	array     JsArrayAccessor[T]
	vm        *goja.Runtime
	converter converter[T]
}

func newJsArray[T any](array JsArrayAccessor[T], vm *goja.Runtime, converter converter[T]) goja.Value {
	nodes := &TrackArray[T]{
		array:     array,
		converter: converter,
		vm:        vm,
	}

	obj := vm.NewDynamicObject(nodes)

	arrayProto := vm.Get("Array").ToObject(vm).Get("prototype").ToObject(vm)
	err := obj.SetPrototype(arrayProto)
	if err != nil {
		panic(err)
	}

	return obj
}

func (a *TrackArray[T]) Get(key string) goja.Value {
	switch key {
	case "length":
		return a.vm.ToValue(a.array.Len())
	case "push":
		return a.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			arg1 := call.Argument(0)
			item := a.converter(arg1)
			a.array.Splice(a.array.Len(), 0, []T{item})
			return goja.Undefined()
		})
	case "pop":
		return a.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if a.array.Len() > 0 {
				a.array.Splice(a.array.Len()-1, 1, nil)
			}

			return goja.Undefined()
		})
	case "shift":
		return a.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if a.array.Len() > 0 {
				a.array.Splice(0, 1, nil)
			}
			return goja.Undefined()
		})
	case "unshift":
		return a.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			values := call.Arguments[0:]
			items := make([]T, 0, len(values))
			for _, v := range values {
				items = append(items, a.converter(v))
			}
			a.array.Splice(0, 0, items)
			return goja.Undefined()
		})
	case "splice":
		return a.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			start := call.Argument(0).ToInteger()
			deleteCount := call.Argument(1).ToInteger()
			values := call.Arguments[2:]
			items := make([]T, 0, len(values))
			for _, v := range values {
				items = append(items, a.converter(v))
			}

			a.array.Splice(int(start), int(deleteCount), items)
			return goja.Undefined()
		})
	}

	if i, err := strconv.Atoi(key); err == nil && i >= 0 && i < a.array.Len() {
		return a.vm.ToValue(a.array.Get(i))
	}

	return nil
}

func (a *TrackArray[T]) Set(key string, val goja.Value) bool {
	if i, err := strconv.Atoi(key); err == nil && i >= 0 {
		a.array.Set(i, a.converter(val))

		return true
	}
	return false
}

func (a *TrackArray[T]) Has(key string) bool {
	if key == "length" {
		return true
	}
	if i, err := strconv.Atoi(key); err == nil {
		return i >= 0 && i < a.array.Len()
	}
	return false
}

func (a *TrackArray[T]) Delete(key string) bool {
	if i, err := strconv.Atoi(key); err == nil && i >= 0 && i < a.array.Len() {
		var zero T
		a.array.Set(i, zero)
		return true
	}
	return false
}

func (a *TrackArray[T]) Keys() []string {
	keys := make([]string, a.array.Len())
	for i := 0; i < a.array.Len(); i++ {
		keys[i] = strconv.Itoa(i)
	}
	keys = append(keys, "length")
	return keys
}
