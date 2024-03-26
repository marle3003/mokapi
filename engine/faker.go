package engine

import (
	"mokapi/engine/common"
	"mokapi/json/generator"
)

type fakerTree struct {
	restore []func() error
	t       *generator.Tree
}

func (ft *fakerTree) Name() string {
	return ft.t.Name
}

func (ft *fakerTree) Test(r *generator.Request) bool {
	return ft.t.Test(r)
}

func (ft *fakerTree) Fake(r *generator.Request) (interface{}, error) {
	return ft.t.Fake(r)
}

func (ft *fakerTree) Append(node common.FakerNode) {
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

func (ft *fakerTree) Insert(index int, node common.FakerNode) error {
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

func (ft *fakerTree) RemoveAt(index int) error {
	return ft.t.RemoveAt(index)
}

func (ft *fakerTree) Remove(name string) error {
	return ft.t.Remove(name)
}

func (ft *fakerTree) Restore() error {
	for _, f := range ft.restore {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}
