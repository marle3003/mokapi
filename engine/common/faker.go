package common

import (
	"mokapi/schema/json/generator"
)

type FakerTree struct {
	restore []func() error
	t       *generator.Tree
}

func NewFakerTree(t *generator.Tree) *FakerTree {
	return &FakerTree{t: t}
}

func (ft *FakerTree) Name() string {
	return ft.t.Name
}

func (ft *FakerTree) Test(r *generator.Request) bool {
	return ft.t.Test(r)
}

func (ft *FakerTree) Fake(r *generator.Request) (interface{}, error) {
	return ft.t.Fake(r)
}

func (ft *FakerTree) Append(node FakerNode) {
	t := &generator.Tree{
		Name:   node.Name(),
		Test:   node.Test,
		Fake:   node.Fake,
		Custom: true,
	}
	ft.t.Append(t)
	ft.restore = append(ft.restore, func() error {
		return ft.t.Remove(t.Name)
	})
}

func (ft *FakerTree) Insert(index int, node FakerNode) error {
	new := &generator.Tree{
		Name:   node.Name(),
		Test:   node.Test,
		Fake:   node.Fake,
		Custom: true,
	}
	err := ft.t.Insert(index, new)
	if err != nil {
		return err
	}
	ft.restore = append(ft.restore, func() error {
		return ft.t.Remove(new.Name)
	})
	return nil
}

func (ft *FakerTree) RemoveAt(index int) error {
	return ft.t.RemoveAt(index)
}

func (ft *FakerTree) Remove(name string) error {
	return ft.t.Remove(name)
}

func (ft *FakerTree) Restore() error {
	for _, f := range ft.restore {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}
