package faker

import (
	"fmt"
	"mokapi/js/util"
	jsonSchema "mokapi/schema/json/schema"
	"reflect"
	"strconv"

	"github.com/dop251/goja"
)

func ToJsonSchema(v goja.Value, rt *goja.Runtime) (*jsonSchema.Schema, error) {
	s := &jsonSchema.Schema{}

	switch v.ExportType().Kind() {
	case reflect.Map:
		break
	case reflect.Bool:
		b := v.ToBoolean()
		return &jsonSchema.Schema{Boolean: &b}, nil
	default:
		return nil, fmt.Errorf("expect JSON schema but got: %v", util.JsType(v.Export()))
	}

	obj := v.ToObject(rt)
	for _, k := range obj.Keys() {
		switch k {
		case "type":
			i := obj.Get(k).Export()
			if arr, ok := i.([]interface{}); ok {
				for _, t := range arr {
					tn, ok := t.(string)
					if !ok {
						return nil, fmt.Errorf("unexpected type for 'type': %v", util.JsType(t))
					}
					s.Type = append(s.Type, tn)
				}
			} else if t, ok := i.(string); ok {
				s.Type = []string{t}
			} else {
				return nil, fmt.Errorf("unexpected type for 'type': %v", util.JsType(i))
			}
		case "enum":
			i := obj.Get(k).Export()
			if enums, ok := i.([]interface{}); ok {
				s.Enum = enums
			} else {
				return nil, fmt.Errorf("unexpected type for 'enum': got %s, expected Array", util.JsType(i))
			}
		case "const":
			c := obj.Get(k).Export()
			s.Const = &c
		case "default":
			s.Default = obj.Get(k).Export()
		case "examples":
			i := obj.Get(k).Export()
			if examples, ok := i.([]interface{}); ok {
				for _, e := range examples {
					s.Examples = append(s.Examples, jsonSchema.Example{Value: e})
				}
			} else {
				return nil, fmt.Errorf("unexpected type for 'enum': got %s, expected Array", util.JsType(i))
			}
		case "multipleOf":
			val := obj.Get(k)
			t := val.ExportType()
			var f float64
			switch t.Kind() {
			case reflect.Float64:
				f = val.ToFloat()
			case reflect.Int64:
				f = float64(val.ToInteger())
			default:
				return nil, fmt.Errorf("unexpected type for 'multipleOf': got %s, expected Number", util.JsType(val))
			}
			s.MultipleOf = &f
		case "maximum":
			val := obj.Get(k)
			t := val.ExportType()
			var f float64
			switch t.Kind() {
			case reflect.Float64:
				f = val.ToFloat()
			case reflect.Int64:
				f = float64(val.ToInteger())
			default:
				return nil, fmt.Errorf("unexpected type for 'maximum': got %s, expected Number", util.JsType(val))
			}
			s.Maximum = &f
		case "exclusiveMaximum":
			val := obj.Get(k)
			kind := val.ExportType().Kind()
			if kind == reflect.Float64 || kind == reflect.Int64 {
				s.ExclusiveMaximum = jsonSchema.NewUnionTypeA[float64, bool](val.ToFloat())
			} else if kind == reflect.Bool {
				s.ExclusiveMaximum = jsonSchema.NewUnionTypeB[float64, bool](val.ToBoolean())
			} else {
				return nil, fmt.Errorf("unexpected type for 'exclusiveMaximum': got %s, expected Number or Boolean", util.JsType(val))
			}
		case "minimum":
			val := obj.Get(k)
			t := val.ExportType()
			var f float64
			switch t.Kind() {
			case reflect.Float64:
				f = val.ToFloat()
			case reflect.Int64:
				f = float64(val.ToInteger())
			default:
				return nil, fmt.Errorf("unexpected type for 'minimum': got %s, expected Number", util.JsType(val))
			}
			s.Minimum = &f
		case "exclusiveMinimum":
			val := obj.Get(k)
			kind := val.ExportType().Kind()
			if kind == reflect.Float64 || kind == reflect.Int64 {
				s.ExclusiveMinimum = jsonSchema.NewUnionTypeA[float64, bool](val.ToFloat())
			} else if kind == reflect.Bool {
				s.ExclusiveMinimum = jsonSchema.NewUnionTypeB[float64, bool](val.ToBoolean())
			} else {
				return nil, fmt.Errorf("unexpected type for 'exclusiveMinimum': got %s, expected Number or Boolean", util.JsType(val))
			}
		case "maxLength":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MaxLength = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'maxLength': got %s, expected Number", util.JsType(i))
			}
		case "minLength":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MinLength = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'minLength': got %s, expected Number", util.JsType(i))
			}
		case "pattern":
			i := obj.Get(k).Export()
			if str, ok := i.(string); ok {
				s.Pattern = str
			} else {
				return nil, fmt.Errorf("unexpected type for 'pattern': got %s, expected String", util.JsType(i))
			}
		case "format":
			i := obj.Get(k).Export()
			if str, ok := i.(string); ok {
				s.Format = str
			} else {
				return nil, fmt.Errorf("unexpected type for 'format': got %s, expected String", util.JsType(i))
			}
		case "items":
			items, err := ToJsonSchema(obj.Get(k), rt)
			if err != nil {
				return nil, err
			}
			s.Items = items
		case "maxItems":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MaxItems = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'maxItems': got %s, expected Integer", util.JsType(i))
			}
		case "minItems":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MinItems = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'minItems': got %s, expected Integer", util.JsType(i))
			}
		case "uniqueItems":
			i := obj.Get(k).Export()
			if b, ok := i.(bool); ok {
				s.UniqueItems = &b
			} else {
				return nil, fmt.Errorf("unexpected type for 'uniqueItems': got %s, expected Boolean", util.JsType(i))
			}
		case "prefixItems":
			val := obj.Get(k)
			if val.ExportType().Kind() != reflect.Slice {
				return nil, fmt.Errorf("unexpected type for 'prefixItems': got %s, expected Array", util.JsType(val))
			}
			arr := val.ToObject(rt)
			length := int(arr.Get("length").ToInteger())
			for i := 0; i < length; i++ {
				item := arr.Get(strconv.Itoa(i))
				pi, err := ToJsonSchema(item, rt)
				if err != nil {
					return nil, err
				}
				s.PrefixItems = append(s.PrefixItems, pi)
			}
		case "contains":
			contains, err := ToJsonSchema(obj.Get(k), rt)
			if err != nil {
				return nil, err
			}
			s.Contains = contains
		case "maxContains":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MaxContains = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'maxContains': got %s, expected Integer", util.JsType(i))
			}
		case "minContains":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MinContains = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'minContains': got %s, expected Integer", util.JsType(i))
			}
		case "x-shuffleItems":
			s.ShuffleItems = obj.Get(k).ToBoolean()
		case "properties":
			s.Properties = &jsonSchema.Schemas{}
			val := obj.Get(k)
			t := val.ExportType()
			if t.Kind() != reflect.Map {
				return nil, fmt.Errorf("unexpected type for 'properties': got %s, expected Object", util.JsType(val))
			}
			propsObj := val.ToObject(rt)
			for _, name := range propsObj.Keys() {
				prop, err := ToJsonSchema(propsObj.Get(name), rt)
				if err != nil {
					return nil, err
				}
				s.Properties.Set(name, prop)
			}
		case "maxProperties":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MaxProperties = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'maxProperties': got %s, expected Integer", util.JsType(i))
			}
		case "minProperties":
			i := obj.Get(k).Export()
			if n64, ok := i.(int64); ok {
				n := int(n64)
				s.MinProperties = &n
			} else {
				return nil, fmt.Errorf("unexpected type for 'minProperties': got %s, expected Integer", util.JsType(i))
			}
		case "patternProperties":
			s.PatternProperties = map[string]*jsonSchema.Schema{}
			val := obj.Get(k)
			t := val.ExportType()
			if t.Kind() != reflect.Map {
				return nil, fmt.Errorf("unexpected type for 'properties': got %s, expected Object", util.JsType(val))
			}
			propsObj := val.ToObject(rt)
			for _, name := range propsObj.Keys() {
				prop, err := ToJsonSchema(propsObj.Get(name), rt)
				if err != nil {
					return nil, err
				}
				s.PatternProperties[name] = prop
			}
		case "required":
			i := obj.Get(k).Export()
			if arr, ok := i.([]interface{}); ok {
				for _, t := range arr {
					req, ok := t.(string)
					if !ok {
						return nil, fmt.Errorf("unexpected type for 'required': %v", util.JsType(t))
					}
					s.Required = append(s.Required, req)
				}
			} else {
				return nil, fmt.Errorf("unexpected type for 'required': got %s, expected Array", util.JsType(i))
			}
		}
	}
	return s, nil
}
