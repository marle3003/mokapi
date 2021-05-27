package runtime

import (
	"fmt"
	"mokapi/providers/workflow/ast"
	"reflect"
	"strings"
)

type selectorVisitor struct {
	outer        *visitor
	resolvedRoot bool
}

func newSelectorVisitor(outer *visitor) *selectorVisitor {
	return &selectorVisitor{outer: outer}
}

func (v *selectorVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		switch t := e.(type) {
		case *ast.Identifier:
			v.outer.stack.Push(t.Name)
		case *ast.Selector:
			if ident, ok := t.X.(*ast.Identifier); ok {
				v.outer.Visit(ident)
				v.Visit(t.Selector)
				return v.Visit(nil)
			}
			return v
		}
		return nil
	}

	selector := v.outer.stack.Pop().(string)
	source := v.outer.stack.Pop()

	m, _ := resolveMember(selector, source)
	v.outer.stack.Push(m)

	return nil
}

func resolveMember(name string, i interface{}) (interface{}, error) {
	if r, ok := i.(ContextResolver); ok {
		return r.Resolve(name)
	}

	v := reflect.ValueOf(i)
	var ptr reflect.Value
	if v.Type().Kind() == reflect.Ptr {
		ptr = v
		v = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(v)
	}

	if v.Kind() == reflect.Map {
		for _, k := range v.MapKeys() {
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("unsupported map key type %q", k.Kind())
			}
			if k.String() == name {
				return v.MapIndex(k).Interface(), nil
			}
		}
	} else if v.Kind() == reflect.Struct {

		fieldName := strings.Title(name)

		f := v.FieldByName(fieldName)
		if !f.IsValid() {
			// check for field on pointer
			f = reflect.Indirect(ptr).FieldByName(fieldName)
		}
		if f.IsValid() {
			return f.Interface(), nil
		}
	}

	return nil, fmt.Errorf("undefined field %q", name)
}
