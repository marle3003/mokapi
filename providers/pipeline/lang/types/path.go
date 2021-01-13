package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
)

type Path struct {
	ObjectImpl
	value     Object
	Parent    *Path
	iterating bool
}

func NewPath(obj Object) *Path {
	return &Path{value: obj}
}

func (p *Path) new(obj Object, iterating bool) *Path {
	return &Path{value: obj, Parent: p, iterating: iterating}
}

func (p *Path) GetField(name string) (Object, error) {
	v, err := p.Resolve(name, nil)
	if err != nil {
		return nil, err
	}
	return p.new(v, false), nil
}

func (p *Path) Set(o Object) error {
	return p.value.Set(o)
}

func (p *Path) InvokeFunc(name string, args map[string]Object) (Object, error) {
	v, err := p.value.InvokeFunc(name, args)
	if err != nil {
		return nil, err
	}
	return p.new(v, false), nil
}

func (p *Path) InvokeOp(op token.Token, o Object) (Object, error) {
	v, err := p.value.InvokeOp(op, o)
	if err != nil {
		return nil, err
	}
	return p.new(v, false), nil
}

func (p *Path) HasField(name string) bool {
	return true
}

func (p *Path) String() string {
	return fmt.Sprintf("%v", p.value)
}

func (p *Path) SetField(name string, v Object) error {
	return p.value.SetField(name, v)
}

func (p *Path) Elem() interface{} {
	return p.value.Elem()
}

//func (p *Path) Resolve(path string, args map[string]Object) (*Path, error) {
//	current := p.value
//	var err error
//	current, err = get(current, path, args)
//	if err != nil {
//		return nil, err
//	}
//	return p.new(current), nil
//}

func (p *Path) Resolve(name string, args map[string]Object) (Object, error) {
	switch name {
	case "*":
		if list, ok := p.value.(Collection); ok {
			return p.new(list.Children(), true), nil
		}
	case "**":
		return p.new(depthFirstIterator(p.value), true), nil
	case "find":
		switch t := p.value.(type) {
		case *Array:
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			for _, item := range t.value {
				if matches, err := match(item); err == nil && matches {
					return p.new(item, false), nil
				}
			}
			return nil, nil
		}
	case "findAll":
		switch t := p.value.(type) {
		case *Array:
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			result := NewArray()
			for _, item := range t.value {
				if matches, err := match(item); err == nil && matches {
					result.Add(item)
				}
			}
			return p.new(result, false), nil
		}
	default:
		if p.iterating {
			switch t := p.value.(type) {
			case *Array:
				a := NewArray()
				for _, o := range t.value {
					r, err := o.GetField(name)
					if err != nil {
						r, err = o.InvokeFunc(name, args)
					}
					if err == nil {
						a.Add(r)
					}
				}
				return p.new(a, false), nil
			}
		}
		r, err := p.value.GetField(name)
		if err != nil {
			r, err = p.value.InvokeFunc(name, args)
		}
		return p.new(r, false), err

	}

	return nil, errors.Errorf("path does not support '%v' on type %v", name, reflect.TypeOf(p.value))
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
		for _, i := range o.children.value {
			depthFirst(i, ch)
			ch <- i
		}
	default:
		ch <- o
	}
}

func newPredicate(c *Closure) Predicate {
	return func(o Object) (bool, error) {
		r, err := c.value([]Object{o})
		if err != nil {
			return false, err
		}

		if p, ok := r.(*Path); ok {
			r = p.value
		}

		if b, ok := r.(*Bool); ok {
			return b.value, nil
		}

		return false, errors.Errorf("unexpected return type: expected bool")
	}
}
