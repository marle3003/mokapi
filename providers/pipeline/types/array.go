package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Array struct {
	value []Object
}

func NewArray() *Array {
	return &Array{value: []Object{}}
}

func (a *Array) Value() interface{} {
	result := make([]interface{}, len(a.value))
	for i, value := range a.value {
		if v, ok := value.(ValueType); ok {
			result[i] = v.Value()
		}
	}
	return result
}

//func (a *Array) GetMember(member string, args []Object) (Object, error) {
//	switch member {
//	case "find":
//		if len(args) != 1 {
//			return nil, fmt.Errorf("syntax error: invalid number of arguments in find")
//		}
//		if closure, ok := args[0].(*Closure); ok {
//			predicate := NewPredicate(closure)
//			return a.Find(predicate)
//		}
//
//		return nil, fmt.Errorf("syntax error: invalid type of argument in find")
//	}
//
//	r := regexp.MustCompile(`\[([0-9]*)\]$`)
//	match := r.FindAllStringSubmatch(member, -1)
//	if len(match) > 0 {
//		index, err := strconv.Atoi(match[0][1])
//		if err != nil {
//			return nil, fmt.Errorf("syntax error: index '%v' is not an integer")
//		}
//		return a.GetIndex(index)
//	}
//	return nil, fmt.Errorf("syntax error: unable to access member '%v' of type array", member)
//}

func (a *Array) GetIndex(index int) (Object, error) {
	if index < len(a.value) {
		return a.value[index], nil
	}

	return nil, fmt.Errorf("syntax error: index '%v' out of range", index)
}

func (a *Array) SetValue(i interface{}) error {
	v, ok := i.([]Object)
	if !ok {
		return fmt.Errorf("syntax error: unable to cast object of type %v to array", reflect.TypeOf(i))
	}
	a.value = v
	return nil
}

func (a *Array) String() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for i, o := range a.value {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(o.String())
	}
	sb.WriteString("]")
	return sb.String()
}

func (a *Array) Operator(op ArithmeticOperator, obj Object) (Object, error) {
	return nil, fmt.Errorf("unsupported operation '%v' on type array", op)
}

func (a *Array) Append(obj Object) {
	a.value = append(a.value, obj)
}

func (a *Array) Find(match Predicate) (Object, error) {
	for _, item := range a.value {
		if matches, err := match(item); err == nil {
			if matches {
				return item, nil
			}
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (a *Array) Equals(obj Object) bool {
	if other, ok := obj.(*Array); ok {
		if len(a.value) != len(other.value) {
			return false
		}

		for i, v := range a.value {
			if !v.(Object).Equals(other.value[i].(Object)) {
				return false
			}
		}
		return true
	}
	return false
}

func (a *Array) GetType() reflect.Type {
	return reflect.TypeOf(a.value)
}
