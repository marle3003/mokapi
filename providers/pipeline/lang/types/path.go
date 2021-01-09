package types

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Path struct {
	ObjectImpl
	value Object
}

func (p *Path) GetField(name string) (Object, error) {
	v, err := get(p.value, name, nil)
	if err != nil {
		return nil, err
	}
	return NewPath(v), nil
}

func (p *Path) HasField(name string) bool {
	return true
}

func NewPath(obj Object) *Path {
	return &Path{value: obj}
}

func (p *Path) String() string {
	return fmt.Sprintf("%v", p.value)
}

func (p *Path) Resolve(path string, args map[string]Object) (*Path, error) {
	current := p.value
	var err error
	current, err = get(current, path, args)
	if err != nil {
		return nil, err
	}
	return NewPath(current), nil
}

func get(obj Object, name string, args map[string]Object) (Object, error) {
	switch name {
	case "*":
		if list, ok := obj.(Collection); ok {
			return list.Children(), nil
		}
	case "**":
		return depthFirstIterator(obj), nil
	case "find":
		switch t := obj.(type) {
		case *Array:
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			for _, item := range t.value {
				if matches, err := match(item); err == nil && matches {
					return item, nil
				}
			}
			return nil, nil
		}
	case "findAll":
		switch t := obj.(type) {
		case *Array:
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			result := NewArray()
			for _, item := range t.value {
				if matches, err := match(item); err == nil && matches {
					result.Add(item)
				}
			}
			return result, nil
		}
	default:
		switch t := obj.(type) {
		case *Array:
			a := NewArray()
			for _, o := range t.value {
				r, err := o.GetField(name)
				if err == nil {
					a.Add(r)
				}
			}
			return a, nil
		default:
			return obj.GetField(name)
		}

	}

	return nil, errors.Errorf("path does not support '%v' on type %v", name, reflect.TypeOf(obj))
}

func depthFirstIterator(obj Object) Object {
	ch := make(chan Object)
	go func() {
		defer close(ch)

		depthFirst(obj, ch)
	}()

	a := NewArray()
	for o := range ch {
		a.Add(o)
	}
	return a
}

func depthFirst(obj Object, ch chan Object) {
	switch o := obj.(type) {
	case *Array:
		for _, i := range o.value {
			depthFirst(i, ch)
			ch <- i
		}
	case *Expando:
		for _, i := range o.value {
			depthFirst(i, ch)
			ch <- i
		}
	case *Node:
		for _, i := range o.children {
			depthFirst(i, ch)
			ch <- i
		}
	default:
		ch <- o
	}
}
