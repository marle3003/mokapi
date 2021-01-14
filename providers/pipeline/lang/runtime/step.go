package runtime

import (
	"mokapi/providers/pipeline/lang/types"
	"reflect"
	"strings"
)

type tagOptions map[string]string

func (v *callVisitor) callStep(step types.Step, args map[string]types.Object) {
	exec := step.Start()
	val := reflect.ValueOf(exec).Elem()
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := getTag(f)
		if tag == "-" {
			continue
		}
		alias, options := parseOption(tag)
		if len(alias) == 0 {
			alias = strings.ToLower(string(f.Name[0])) + f.Name[1:]
		}

		var value types.Object
		if arg, ok := args[alias]; ok {
			value = arg
		} else {
			position, positionExists := options["position"]
			if arg, ok := args[position]; positionExists && ok {
				value = arg
			} else if _, required := options["required"]; required {
				v.outer.AddErrorf(v.call.Pos(), "missing required argument '%v'", alias)
				return
			} else {
				continue
			}
		}

		field := val.FieldByName(f.Name)
		fieldValue, err := types.ConvertFrom(value, field.Type())
		if err != nil {
			v.outer.AddError(v.call.Pos(), err.Error())
			return
		}
		field.Set(fieldValue)
	}

	result, err := exec.Run(v.outer.Scope())
	if err != nil {
		v.outer.AddError(v.call.Pos(), err.Error())
		return
	}

	obj, err := types.Convert(result)
	if err != nil {
		v.outer.AddError(v.call.Pos(), err.Error())
	} else {
		v.outer.Stack().Push(obj)
	}
}

func parseOption(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	options := tagOptions{}
	for _, opt := range s[1:] {
		kv := strings.Split(opt, "=")
		if len(kv) > 1 {
			options[kv[0]] = kv[1]
		} else {
			options[kv[0]] = ""
		}
	}
	return s[0], options
}

func getTag(f reflect.StructField) string {
	tag := f.Tag.Get("step")
	if len(tag) > 0 {
		return tag
	}
	tag = string(f.Tag)
	if strings.Index(tag, ":") < 0 {
		return tag
	}
	return ""
}
