package common

import (
	"mokapi/schema/json/generator"
)

type FakerTree struct {
	restore []func() error
	t       *generator.Node
}

func NewFakerTree(t *generator.Node) *FakerTree {
	return &FakerTree{t: t}
}

func (ft *FakerTree) Name() string {
	return ft.t.Name
}

func (ft *FakerTree) Fake(r *generator.Request) (interface{}, error) {
	return ft.t.Fake(r)
}

func (ft *FakerTree) Append(node FakerNode) {
	t := &generator.Node{
		Name:   node.Name(),
		Fake:   node.Fake,
		Custom: true,
	}
	ft.t.Append(t)
	ft.restore = append(ft.restore, func() error {
		return ft.t.Remove(t.Name)
	})
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
