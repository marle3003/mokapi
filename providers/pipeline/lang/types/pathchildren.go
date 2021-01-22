package types

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
	"strconv"
)

type PathChildren struct {
	values          []Path
	childTraversing bool
	Parent          Path
}

func newPathChildren(list Collection, parent Path) *PathChildren {
	p := &PathChildren{Parent: parent, childTraversing: true}
	for _, i := range list.Children().value {
		p.values = append(p.values, newPathFromParent(i, parent))
	}
	return p
}

func (p *PathChildren) Value() Object {
	a := NewArray()
	for _, i := range p.values {
		a.Add(i.Value())
	}
	return a
}

func (p *PathChildren) GetField(name string) (Object, error) {
	return p.Resolve(name, nil)
}

func (p *PathChildren) Set(o Object) error {
	if !p.childTraversing {
		return errors.Errorf("invalid operation on path collection")
	}
	for _, i := range p.values {
		if err := i.Set(o); err != nil {
			return err
		}
	}
	return nil
}

func (p *PathChildren) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return p.Resolve(name, args)
}

func (p *PathChildren) InvokeOp(op token.Token, o Object) (Object, error) {
	if !p.childTraversing {
		return nil, errors.Errorf("invalid operation %v on path collection", op)
	}
	var values []Path
	for _, i := range p.values {
		if r, err := i.InvokeOp(op, o); err != nil {
			return nil, err
		} else {
			values = append(values, NewPath(r))
		}
	}
	return &PathChildren{values: values}, nil
}

func (p *PathChildren) HasField(name string) bool {
	return true
}

func (p *PathChildren) String() string {
	var values []string
	for _, i := range p.values {
		values = append(values, i.String())
	}
	return fmt.Sprintf("%v", values)
}

func (p *PathChildren) SetField(name string, v Object) error {
	if !p.childTraversing {
		return errors.Errorf("invalid operation on path collection")
	}
	for _, i := range p.values {
		if err := i.SetField(name, v); err != nil {
			return err
		}
	}
	return nil
}

func (p *PathChildren) Elem() interface{} {
	var a []interface{}
	for _, i := range p.values {
		a = append(a, i.Elem())
	}
	return a
}

func (p *PathChildren) GetType() reflect.Type {
	return reflect.TypeOf(p)
}

func (p *PathChildren) Resolve(name string, args map[string]Object) (Path, error) {
	switch name {
	case "..":
		if p.childTraversing {
			return p.iterate(func(p Path) (Path, error) {
				f, err := p.GetField(name)
				if err != nil {
					var o Object
					o, err = p.InvokeFunc(name, args)
					if err != nil {
						return nil, err
					}
					return newPathFromParent(o, p), nil
				}
				return newPathFromParent(f, p), nil
			})
		} else {
			return p.Parent, nil
		}
	case "*":
		p.childTraversing = true
		return p, nil
	case "**":
		return &PathChildren{values: p.depthFirstIterator(), Parent: p, childTraversing: true}, nil
	case "find":
		if p.childTraversing {
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			for _, path := range p.values {
				if matches, err := match(path.Value()); err == nil && matches {
					return p, nil
				}
			}
		}
		return nil, nil
	case "findAll":
		if p.childTraversing {
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			var values []Path
			for _, path := range p.values {
				if matches, err := match(path.Value()); err == nil && matches {
					values = append(values, path)
				}
			}
			return &PathChildren{values: values, Parent: p}, nil
		}
		return &PathChildren{Parent: p}, nil
	case "any":
		if p.childTraversing {
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			for _, path := range p.values {
				if matches, err := match(path.Value()); err == nil && matches {
					return newPath(NewBool(true)), nil
				}
			}
			return newPath(NewBool(false)), nil
		}
		return &PathChildren{Parent: p}, nil
	case "every":
		if p.childTraversing {
			closure := args["0"].(*Closure)
			match := newPredicate(closure)
			for _, path := range p.values {
				if matches, err := match(path.Value()); err != nil || !matches {
					return newPath(NewBool(false)), nil
				}
			}
			return newPath(NewBool(true)), nil
		}
		return &PathChildren{Parent: p}, nil
	default:
		if p.childTraversing {
			return p.iterate(func(i Path) (Path, error) {
				r, err := i.GetField(name)
				if err != nil {
					r, err = i.InvokeFunc(name, args)
				}
				return newPathFromParent(r, i), nil
			})
		} else if index, err := strconv.Atoi(name); err == nil {
			return p.values[index], nil
		}

	}

	return nil, errors.Errorf("path does not support '%v' on type %v", name, reflect.TypeOf(p))
}

func (p *PathChildren) iterate(action func(Path) (Path, error)) (Path, error) {
	it, err := p.newIterator()
	if err != nil {
		return nil, err
	}
	var v []Path
	for o := range it {
		r, err := action(o)
		if err != nil {
			log.Debugf("ignored error during iteration over path: %v", err.Error())
		} else {
			v = append(v, r)
		}
	}
	return &PathChildren{values: v, Parent: p}, nil
}

func (p *PathChildren) newIterator() (ch chan Path, err error) {
	ch = make(chan Path)

	go func() {
		defer close(ch)
		for _, i := range p.values {
			ch <- i
		}
	}()
	return
}

func (p *PathChildren) depthFirstIterator() []Path {
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

func (p *PathChildren) depthFirst(ch chan Path) {
	for _, v := range p.values {
		for _, i := range v.depthFirstIterator() {
			ch <- i
		}
	}
}
