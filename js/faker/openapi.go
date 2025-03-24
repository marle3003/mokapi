package faker

import (
	"fmt"
	"github.com/dop251/goja"
	"mokapi/js/util"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"reflect"
)

func isOpenApiSchema(o *goja.Object) bool {
	if v := o.Get("xml"); v != nil {
		return true
	}
	if v := o.Get("example"); v != nil {
		return true
	}

	if schemaDef := o.Get("$schema"); schemaDef != nil && schemaDef.ExportType().Kind() == reflect.String {
		def := schemaDef.String()
		switch def {
		case "https://spec.openapis.org/oas/3.1/dialect/base":
			return true
		}
	}

	return false
}

func ToOpenAPISchema(v goja.Value, rt *goja.Runtime) (*schema.Schema, error) {
	s := &schema.Schema{SubSchema: &schema.SubSchema{}}

	if v == nil {
		return nil, nil
	}

	if v.ExportType().Kind() != reflect.Map {
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
				return nil, fmt.Errorf("unexpected type for 'enum'")
			}
		case "const":
			c := obj.Get(k).Export()
			s.Const = &c
		case "default":
			s.Default = obj.Get(k).Export()
		case "example":
			i := obj.Get(k).Export()
			s.Example = &jsonSchema.Example{Value: i}
		case "examples":
			i := obj.Get(k).Export()
			if examples, ok := i.([]interface{}); ok {
				for _, e := range examples {
					s.Examples = append(s.Examples, jsonSchema.Example{Value: e})
				}
			} else {
				return nil, fmt.Errorf("unexpected type for 'examples'")
			}
		case "multipleOf":
			f := obj.Get(k).ToFloat()
			s.MultipleOf = &f
		case "maximum":
			f := obj.Get(k).ToFloat()
			s.Maximum = &f
		case "exclusiveMaximum":
			val := obj.Get(k)
			kind := val.ExportType().Kind()
			if kind == reflect.Float64 || kind == reflect.Int64 {
				s.ExclusiveMaximum = jsonSchema.NewUnionTypeA[float64, bool](val.ToFloat())
			} else if kind == reflect.Bool {
				s.ExclusiveMaximum = jsonSchema.NewUnionTypeB[float64, bool](val.ToBoolean())
			} else {
				return nil, fmt.Errorf("unexpected type for 'exclusiveMaximum': %v", util.JsType(val.Export()))
			}
		case "minimum":
			f := obj.Get(k).ToFloat()
			s.Minimum = &f
		case "exclusiveMinimum":
			val := obj.Get(k)
			kind := val.ExportType().Kind()
			if kind == reflect.Float64 || kind == reflect.Int64 {
				s.ExclusiveMinimum = jsonSchema.NewUnionTypeA[float64, bool](val.ToFloat())
			} else if kind == reflect.Bool {
				s.ExclusiveMinimum = jsonSchema.NewUnionTypeB[float64, bool](val.ToBoolean())
			} else {
				return nil, fmt.Errorf("unexpected type for 'exclusiveMinimum': %v", util.JsType(val.Export()))
			}
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
			items, err := ToOpenAPISchema(obj.Get(k), rt)
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
		//	i := int(obj.Get(k).ToInteger())
		//	s.MaxContains = &i
		case "minContains":
		//	i := int(obj.Get(k).ToInteger())
		//	s.MinContains = &i
		case "x-shuffleItems":
			s.ShuffleItems = obj.Get(k).ToBoolean()
		case "properties":
			s.Properties = &schema.Schemas{}
			propsObj := obj.Get(k).ToObject(rt)
			for _, name := range propsObj.Keys() {
				prop, err := ToOpenAPISchema(propsObj.Get(name), rt)
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
					s.Required = append(s.Type, req)
				}
			} else {
				return nil, fmt.Errorf("unexpected type for 'required': %v", util.JsType(i))
			}
		case "xml":
			xml := &schema.Xml{}
			s.Xml = xml
			o := obj.Get(k).ToObject(rt)
			for _, name := range o.Keys() {
				val := o.Get(name)
				switch name {
				case "wrapped":
					xml.Wrapped = val.ToBoolean()
				case "name":
					xml.Name = val.String()
				case "attribute":
					xml.Attribute = val.ToBoolean()
				case "prefix":
					xml.Prefix = val.String()
				case "namespace":
					xml.Namespace = val.String()
				}
			}
		}
	}
	return s, nil
}
