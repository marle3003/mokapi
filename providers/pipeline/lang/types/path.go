package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Path struct {
	ObjectImpl
	value interface{}
}

func (p *Path) GetField(name string) (Object, error) {
	v, err := get(p.value, name)
	if err != nil {
		return nil, err
	}
	return NewPath(v), nil
}

func NewPath(obj interface{}) *Path {
	return &Path{value: obj}
}

func (p *Path) String() string {
	return fmt.Sprintf("%v", p.value)
}

func (p *Path) Resolve(segments []string) (*Path, error) {
	current := p.value
	for _, s := range segments {
		var err error
		current, err = get(current, s)
		if err != nil {
			return nil, err
		}
	}
	return NewPath(current), nil
}

func get(obj interface{}, name string) (interface{}, error) {
	switch name {
	case "*":
		if list, ok := obj.(Collection); ok {
			return list.Children(), nil
		}
	//case "**":
	//	if list, ok := obj.(Collection); ok {
	//		return list.depthFirst(), nil
	//	}
	default:
		switch t := obj.(type) {
		case *Array:
			a := NewArray()
			for _, o := range t.value {
				r, err := getField(o, name)
				if err == nil {
					a.Add(r)
				}
			}
			return a, nil
		default:
			return getField(obj, name)
		}

	}

	return nil, errors.Errorf("unsupported type %v", reflect.TypeOf(obj))
}
