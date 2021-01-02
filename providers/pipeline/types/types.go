package types

import (
	"fmt"
	"reflect"
)

type ArithmeticOperator string

const (
	Addition       ArithmeticOperator = "+"
	Subtraction    ArithmeticOperator = "-"
	Multiplication ArithmeticOperator = "*"
	Division       ArithmeticOperator = "/"
	Remainder      ArithmeticOperator = "%"
)

type Object interface {
	String() string
	Equals(obj Object) bool
	GetType() reflect.Type
}

//type Object interface{
//	Value() interface{}
//	Set(interface{}) error
//	GetMember(member string, args []Object) (Object, error)
//	Process(operator string, value Object) (Object, error)
//	String() string
//}

type ValueType interface {
	Value() interface{}
	SetValue(interface{}) error
	Operator(op ArithmeticOperator, obj Object) (Object, error)
}

type Comparable interface {
	CompareTo(obj Object) (int, error)
}

type Predicate func(Object) (bool, error)

type Collection interface {
	Add(obj Object)
	Find(match Predicate) (Object, error)
	GetEnumerator() []Object
	//FindAll() []Type
}

type Dictionary interface {
	Get(name string)
}

type Class interface {
	Invoke(name string, args []Object) (Object, error)
	Set(name string, obj Object) error
}

type ClosureFunc func(parameters []Object) (Object, error)

func Convert(i interface{}) (Object, error) {
	if obj, ok := i.(Object); ok {
		return obj, nil
	}

	if i == nil {
		return nil, nil
	}
	switch v := i.(type) {
	case int:
		return NewNumber(float64(v)), nil
	case float64:
		return NewNumber(v), nil
	case float32:
		return NewNumber(float64(v)), nil
	case string:
		return NewString(v), nil
	case []interface{}:
		a := NewArray()
		for _, e := range v {
			o, err := Convert(e)
			if err != nil {
				return nil, err
			}
			a.Append(o)
		}
		return a, nil
	case interface{}:
		return NewReference(i), nil
	}
	return nil, fmt.Errorf("unable to convert type '%v'", reflect.TypeOf(i))
}

func ConvertFrom(obj Object, t reflect.Type) (reflect.Value, error) {
	if obj == nil {
		return reflect.Zero(t), nil
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(obj.String()), nil
	case reflect.Int:
		switch arg := obj.(type) {
		case *Number:
			return reflect.ValueOf(int(arg.value)), nil
		default:
			return reflect.New(t), fmt.Errorf("unable to cast type %v to int", obj.GetType())
		}
	case reflect.Float64:
		switch arg := obj.(type) {
		case *Number:
			return reflect.ValueOf(arg.value), nil
		default:
			return reflect.New(t), fmt.Errorf("unable to cast type %v to float", obj.GetType())
		}
	case reflect.Interface:
		switch arg := obj.(type) {
		case ValueType:
			{
				return reflect.ValueOf(arg.Value()), nil
			}
		}
	case reflect.Func:
		switch arg := obj.(type) {
		case *Closure:
			v := reflect.ValueOf(arg.value)
			// create a function which calls closure function with the given parameters
			fn := func(args []reflect.Value) []reflect.Value {
				results := make([]reflect.Value, t.NumOut())
				in := make([]Object, t.NumIn())
				// converts the given parameters to a slice of types.Object
				for i := range in {
					obj, err := Convert(args[i].Interface())
					if err != nil {
						panic(err)
					}
					in[i] = obj
				}

				// call the closure function
				values := v.Call([]reflect.Value{reflect.ValueOf(in)})

				// returning result values: (types.Object, error)
				if len(results) > 0 {
					i := values[0].Interface()
					if i != nil {
						v, err := ConvertFrom(i.(Object), t.Out(0))
						if err != nil {
							panic(err)
						}
						results[0] = v
					} else {
						results[0] = reflect.Zero(t.Out(0))
					}
				}
				if len(results) > 1 {
					results[1] = values[1]
				}

				return results
			}
			ins, outs := make([]reflect.Type, t.NumIn()), make([]reflect.Type, t.NumOut())
			for i := 0; i < t.NumIn(); i++ {
				ins[i] = t.In(i)
			}
			for i := 0; i < t.NumOut(); i++ {
				outs[i] = t.Out(i)
			}
			// creating the function
			return reflect.MakeFunc(reflect.FuncOf(ins, outs, false), fn), nil
		}
	default:
		switch arg := obj.(type) {
		case ValueType:
			return reflect.ValueOf(arg.Value()), nil
		}
	}

	return reflect.New(t), fmt.Errorf("unsupported paramter type '%v'", t.Kind())
}
