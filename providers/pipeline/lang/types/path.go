package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
)

type Path interface {
	Object
	Resolve(name string, args map[string]Object) (Path, error)
	Value() Object
	depthFirstIterator() []Path
}

func newPath(target Object) Path {
	return &PathValue{value: target}
}

type PathValue struct {
	value  Object
	Parent Path
}

func NewPath(obj Object) *PathValue {
	return &PathValue{value: obj}
}

func newPathFromParent(obj Object, parent Path) *PathValue {
	return &PathValue{value: obj, Parent: parent}
}

func (p *PathValue) new(obj Object, iterating bool) *PathValue {
	return &PathValue{value: obj, Parent: p}
}

func (p *PathValue) Value() Object {
	return p.value
}

func (p *PathValue) GetField(name string) (Object, error) {
	v, err := p.Resolve(name, nil)
	if err != nil {
		return nil, err
	}
	return p.new(v, false), nil
}

func (p *PathValue) Set(o Object) error {
	return p.value.Set(o)
}

func (p *PathValue) InvokeFunc(name string, args map[string]Object) (Object, error) {
	v, err := p.value.InvokeFunc(name, args)
	if err != nil {
		return nil, err
	}
	return p.new(v, false), nil
}

func (p *PathValue) InvokeOp(op token.Token, o Object) (Object, error) {
	v, err := p.value.InvokeOp(op, o)
	if err != nil {
		return nil, err
	}
	return p.new(v, false), nil
}

func (p *PathValue) HasField(name string) bool {
	return true
}

func (p *PathValue) String() string {
	return fmt.Sprintf("%v", p.value)
}

func (p *PathValue) SetField(name string, v Object) error {
	return p.value.SetField(name, v)
}

func (p *PathValue) Elem() interface{} {
	return p.value.Elem()
}

func (p *PathValue) GetType() reflect.Type {
	return reflect.TypeOf(p.value)
}

func (p *PathValue) Resolve(name string, args map[string]Object) (Path, error) {
	switch name {
	case "..":
		return p.Parent, nil
	case "*":
		if list, ok := p.value.(Collection); ok {
			return newPathChildren(list, p), nil
		} else {
			return nil, errors.Errorf("path element '%v' is not a collection", p.value.GetType())
		}
	case "@*":
		f, err := p.value.GetField(name)
		if err != nil {
			return nil, err
		} else if a, isArray := f.(*Array); isArray {
			return newPathFromParent(a, p), nil
		}
		return &PathValue{value: f, Parent: p}, nil
	case "**":
		return &PathChildren{values: p.depthFirstIterator(), Parent: p, childTraversing: true}, nil
	default:
		r, err := p.value.GetField(name)
		if err != nil {
			r, err = p.value.InvokeFunc(name, args)
			if err != nil {
				return nil, errors.Errorf("field or func '%v' not found on type %v", name, reflect.TypeOf(p.value))
			}
		}
		return &PathValue{value: r, Parent: p}, err
	}
}

func (p *PathValue) depthFirstIterator() []Path {
	ch := make(chan Path)
	go func() {
		defer close(ch)

		p.depthFirst(ch)
	}()

	var values []Path
	for path := range ch {
		values = append(values, path)
	}
	return values
}

func (p *PathValue) depthFirst(ch chan Path) {
	switch o := p.value.(type) {
	case *Array:
		for _, i := range o.value {
			path := &PathValue{value: i, Parent: p}
			path.depthFirst(ch)
			ch <- path
		}
	case *Expando:
		for _, i := range o.value {
			path := &PathValue{value: i, Parent: p}
			path.depthFirst(ch)
			ch <- path
		}
	case *Node:
		for _, i := range o.attributes.value {
			path := &PathValue{value: i, Parent: p}
			path.depthFirst(ch)
			ch <- path
		}
		for _, i := range o.children.value {
			path := &PathValue{value: i, Parent: p}
			path.depthFirst(ch)
			ch <- path
		}
		ch <- &PathValue{value: NewString(o.content), Parent: p}
	}
	ch <- p
}

func newPredicate(c *Closure) Predicate {
	return func(o Object) (bool, error) {
		r, err := c.value([]Object{o})
		if err != nil {
			return false, err
		}

		if p, ok := r.(Path); ok {
			r = p.Value()
		}

		if b, ok := r.(*Bool); ok {
			return b.value, nil
		}

		return false, errors.Errorf("unexpected return type: expected bool")
	}
}
