package faker

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/js/util"
	jsonSchema "mokapi/schema/json/schema"
	"reflect"
)

func ToJsonSchema(v goja.Value, rt *goja.Runtime) (*jsonSchema.Schema, error) {
	s := &jsonSchema.Schema{}

	if v.ExportType().Kind() != reflect.Map {
		return nil, fmt.Errorf("expect JSON schema but got: %T", v.Export())
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
						return nil, fmt.Errorf("unexpected type: %v", util.JsType(t))
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
				return nil, fmt.Errorf("unexpected type for 'enum': %v", util.JsType(i))
			}
		case "const":
			c := obj.Get(k).Export()
			s.Const = &c
		case "default":
			s.Default = obj.Get(k).Export()
		case "examples":
			i := obj.Get(k).Export()
			if examples, ok := i.([]interface{}); ok {
				for _, item := range examples {
					s.Examples = append(s.Examples, jsonSchema.Example{Value: item})
				}
			} else {
				return nil, fmt.Errorf("unexpected type for 'examples': %v", util.JsType(i))
			}
		case "multipleOf":
			f := obj.Get(k).ToFloat()
			s.MultipleOf = &f
		case "maximum":
			f := obj.Get(k).ToFloat()
			s.Maximum = &f
		case "exclusiveMaximum":
			ex := obj.Get(k)
			kind := ex.ExportType().Kind()
			if kind != reflect.Float64 && kind != reflect.Int64 {
				return nil, fmt.Errorf("unexpected type for 'exclusiveMaximum': %v", util.JsType(ex.Export()))
			}
			f := obj.Get(k).ToFloat()
			s.ExclusiveMaximum = jsonSchema.NewUnionTypeA[float64, bool](f)
		case "minimum":
			f := obj.Get(k).ToFloat()
			s.Minimum = &f
		case "exclusiveMinimum":
			ex := obj.Get(k)
			kind := ex.ExportType().Kind()
			if kind != reflect.Float64 && kind != reflect.Int64 {
				return nil, fmt.Errorf("unexpected type for 'exclusiveMinimum': %v", util.JsType(ex.Export()))
			}
			f := ex.ToFloat()
			s.ExclusiveMinimum = jsonSchema.NewUnionTypeA[float64, bool](f)
		case "maxLength":
			i := int(obj.Get(k).ToInteger())
			s.MaxLength = &i
		case "minLength":
			i := int(obj.Get(k).ToInteger())
			s.MinLength = &i
		case "pattern":
			s.Pattern = obj.Get(k).String()
		case "format":
			s.Format = obj.Get(k).String()
		case "items":
			items, err := ToJsonSchema(obj.Get(k), rt)
			if err != nil {
				return nil, err
			}
			s.Items = items
		case "maxItems":
			i := int(obj.Get(k).ToInteger())
			s.MaxItems = &i
		case "minItems":
			i := int(obj.Get(k).ToInteger())
			s.MinItems = &i
		case "uniqueItems":
			s.UniqueItems = obj.Get(k).ToBoolean()
		case "maxContains":
			i := int(obj.Get(k).ToInteger())
			s.MaxContains = &i
		case "minContains":
			i := int(obj.Get(k).ToInteger())
			s.MinContains = &i
		case "x-shuffleItems":
			s.ShuffleItems = obj.Get(k).ToBoolean()
		case "properties":
			s.Properties = &jsonSchema.Schemas{}
			propsObj := obj.Get(k).ToObject(rt)
			for _, name := range propsObj.Keys() {
				prop, err := ToJsonSchema(propsObj.Get(name), rt)
				if err != nil {
					return nil, err
				}
				s.Properties.Set(name, prop)
			}
		case "maxProperties":
			i := int(obj.Get(k).ToInteger())
			s.MaxProperties = &i
		case "minProperties":
			i := int(obj.Get(k).ToInteger())
			s.MinProperties = &i
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
				return nil, fmt.Errorf("unexpected type for 'required': %v", util.JsType(i))
			}
		}
	}
	return s, nil
}
