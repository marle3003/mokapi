package mokapi

import (
	"fmt"
	"mokapi/lib"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

// Proxy provides an isolated, runtime-bound copy of any JavaScript value coming from Goja.
// Its purpose is to avoid side effects when JavaScript event handlers or user scripts mutate objects.
type Proxy struct {
	target        reflect.Value
	wasPointer    bool
	vm            *goja.Runtime
	KeyNormalizer func(string) string
	ToJSValue     func(vm *goja.Runtime, k string, v any) goja.Value
}

func NewProxy(target any, vm *goja.Runtime) *Proxy {
	v := reflect.ValueOf(target)
	return newProxy(v, vm)
}

func newProxy(v reflect.Value, vm *goja.Runtime) *Proxy {
	wasPointer := v.Kind() == reflect.Ptr
	if v.Kind() == reflect.Interface {
		ptr := reflect.New(v.Elem().Type())
		ptr.Elem().Set(v.Elem())
		v = ptr
		wasPointer = false
	}
	if v.Kind() != reflect.Ptr {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		v = ptr
	}

	return &Proxy{target: v, vm: vm, wasPointer: wasPointer}
}

func (p *Proxy) Get(key string) goja.Value {
	if !p.target.IsValid() {
		return goja.Undefined()
	}

	target := unwrap(p.target)
	switch target.Kind() {
	case reflect.Map:
		key = p.normalizeKey(key)
		v := target.MapIndex(reflect.ValueOf(key))
		return p.toJSValue(key, v)
	case reflect.Struct:
		f := getField(target, key, "json")
		if f.IsValid() {
			return p.toJSValue(key, f)
		}
	case reflect.Slice:
		switch key {
		case "length":
			return p.vm.ToValue(target.Len())
		case "push":
			return p.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				arg1 := call.Argument(0)
				item := arg1.Export()
				splice(p.target, reflect.ValueOf([]any{item}), target.Len(), 0)
				return goja.Undefined()
			})
		case "pop":
			return p.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if target.Len() > 0 {
					v := target.Index(target.Len() - 1)
					splice(p.target, reflect.Value{}, target.Len()-1, 1)
					return p.vm.NewDynamicObject(newProxy(v, p.vm))
				}

				return goja.Undefined()
			})
		case "shift":
			return p.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if target.Len() > 0 {
					v := target.Index(0)
					splice(target, reflect.Value{}, 0, 1)
					return p.vm.NewDynamicObject(newProxy(v, p.vm))
				}
				return goja.Undefined()
			})
		case "unshift":
			return p.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				values := call.Arguments[0:]
				items := make([]any, 0, len(values))
				for _, v := range values {
					items = append(items, NewProxy(v.Export(), p.vm))
				}
				splice(target, reflect.ValueOf(items), 0, 0)
				return goja.Undefined()
			})
		case "splice":
			return p.vm.ToValue(func(call goja.FunctionCall) goja.Value {
				start := call.Argument(0).ToInteger()
				deleteCount := call.Argument(1).ToInteger()
				values := call.Arguments[2:]
				items := make([]any, 0, len(values))
				for _, v := range values {
					items = append(items, NewProxy(v.Export(), p.vm))
				}

				splice(target, reflect.ValueOf(items), int(start), int(deleteCount))
				return goja.Undefined()
			})
		default:
			if i, err := strconv.Atoi(key); err == nil && i >= 0 && i < target.Len() {
				v := target.Index(i).Interface()
				return p.vm.ToValue(v)
			}
			return goja.Undefined()
		}
	}
	panic(fmt.Sprintf("%s is not defined", key))
}

func (p *Proxy) Has(key string) bool {
	if !p.target.IsValid() {
		return false
	}

	target := unwrap(p.target)
	switch target.Kind() {
	case reflect.Map:
		key = p.normalizeKey(key)
		k := target.MapIndex(reflect.ValueOf(key))
		return k.IsValid()
	case reflect.Struct:
		f := getField(target, key, "json")
		return f.IsValid()
	default:
		return false
	}
}

func (p *Proxy) Set(key string, value goja.Value) bool {
	if !p.target.IsValid() {
		return false
	}
	target := unwrap(p.target)
	switch target.Kind() {
	case reflect.Map:
		key = p.normalizeKey(key)
		v, err := convertTo(target.Type().Elem(), reflect.ValueOf(value.Export()))
		if err != nil {
			panic(p.vm.ToValue(err))
		}
		target.SetMapIndex(reflect.ValueOf(key), v)
		return true
	case reflect.Struct:
		f := getField(target, key, "json")
		err := assignValue(f, value.Export(), key)
		if err != nil {
			panic(p.vm.ToValue(err))
		}
		return true
	default:
		return false
	}
}

func (p *Proxy) Delete(key string) bool {
	if !p.target.IsValid() {
		return false
	}
	if p.target.Kind() == reflect.Map {
		p.target.SetMapIndex(reflect.ValueOf(key), reflect.Value{})
		return true
	}
	return false
}

func (p *Proxy) Keys() []string {
	var result []string
	target := p.target
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	if target.Kind() == reflect.Map {
		for _, k := range target.MapKeys() {
			result = append(result, k.Interface().(string))
		}
	}
	if target.Kind() == reflect.Struct {
		t := target.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			v := f.Tag.Get("json")
			if v != "" {
				tagValues := strings.Split(v, ",")
				result = append(result, tagValues[0])
			} else {
				result = append(result, f.Name)
			}
		}
	}

	return result
}

func (p *Proxy) normalizeKey(key string) string {
	if p.KeyNormalizer != nil {
		return p.KeyNormalizer(key)
	}
	return key
}

func (p *Proxy) toJSValue(key string, v reflect.Value) goja.Value {
	if !v.IsValid() {
		return goja.Undefined()
	}

	if p.ToJSValue != nil {
		return p.ToJSValue(p.vm, key, v.Interface())
	}
	return p.vm.NewDynamicObject(newProxy(v, p.vm))
}

func (p *Proxy) Export() any {
	var v any
	if p.wasPointer {
		v = p.target.Interface()
	} else {
		v = p.target.Elem().Interface()
	}
	switch vv := v.(type) {
	case []any:
		for i, e := range vv {
			vv[i] = Export(e)
		}
	case map[string]any:
		for key, val := range vv {
			vv[key] = Export(val)
		}
	}
	return Export(v)
}

func getField(structValue reflect.Value, name, tag string) reflect.Value {
	name = capitalize(name)
	for i := 0; i < structValue.NumField(); i++ {
		f := structValue.Type().Field(i)
		if f.Name == name {
			return structValue.Field(i)
		}
		t := f.Tag.Get(tag)
		tagValues := strings.Split(t, ",")
		for _, tagValue := range tagValues {
			if tagValue == name {
				return structValue.Field(i)
			}
		}
	}
	return reflect.Value{}
}

func splice(slice, toAdd reflect.Value, start int, deleteCount int) {
	if slice.Kind() == reflect.Ptr {
		splice(slice.Elem(), toAdd, start, deleteCount)
		return
	}

	if start < 0 {
		return
	}

	v := slice
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	end := start + deleteCount
	if end > v.Len() {
		end = v.Len()
	}

	n := v.Len()
	if toAdd.IsValid() {
		n += toAdd.Len()
	}
	result := reflect.MakeSlice(v.Type(), 0, n-deleteCount)

	// Append array[:start]
	result = reflect.AppendSlice(result, v.Slice(0, start))
	if toAdd.IsValid() && toAdd.Len() > 0 {
		// Append toAdd
		result = reflect.AppendSlice(result, toAdd)
	}
	// Append array[end:]
	result = reflect.AppendSlice(result, v.Slice(end, v.Len()))

	slice.Set(result)
}

func assignValue(field reflect.Value, value any, fieldName string) error {
	v, err := convertTo(field.Type(), reflect.ValueOf(value))
	if err != nil {
		return fmt.Errorf("failed to set %s: %w", fieldName, err)
	}
	field.Set(v)
	return nil
}

func convertTo(targetType reflect.Type, value reflect.Value) (reflect.Value, error) {
	if targetType.Kind() == reflect.Interface && value.Kind() != reflect.Ptr {
		ptr := reflect.New(value.Type())
		ptr.Elem().Set(value)
		value = ptr
	}

	sourceType := value.Type()
	if sourceType.AssignableTo(targetType) {
		return value, nil
	}

	if sourceType.ConvertibleTo(targetType) {
		return value.Convert(targetType), nil
	}

	if targetType.Kind() == reflect.Slice && sourceType.Kind() == reflect.Slice {
		result := reflect.MakeSlice(targetType, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			if item.Type().AssignableTo(targetType.Elem()) {
				result = reflect.Append(result, item)
			} else if sourceType.ConvertibleTo(targetType) {
				result = reflect.Append(result, item.Convert(targetType))
			} else {
				return reflect.Value{}, fmt.Errorf("items must be a %s", lib.TypeString(targetType.Elem()))
			}
		}
		return result, nil
	}

	if targetType.Kind() == reflect.Map && sourceType.Kind() == reflect.Map {
		result := reflect.MakeMap(targetType)
		for _, key := range value.MapKeys() {
			item, err := convertTo(targetType.Elem(), value.MapIndex(key))
			if err != nil {
				return reflect.Value{}, err
			}
			result.SetMapIndex(key, item)
		}
		return result, nil
	}

	return reflect.Value{}, fmt.Errorf("expected %s but got %s", lib.TypeString(targetType), lib.TypeString(sourceType))
}

func unwrap(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			if v.IsNil() {
				return v
			}
			v = v.Elem()

		default:
			return v
		}
	}
}

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}
