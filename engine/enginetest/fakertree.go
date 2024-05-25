package enginetest

import (
	"mokapi/engine/common"
	"mokapi/schema/json/generator"
)

type FakerTree struct {
	Tree *generator.Tree
}

func (ft *FakerTree) Name() string {
	return ft.Tree.Name
}

func (ft *FakerTree) Test(r *generator.Request) bool {
	return ft.Tree.Test(r)
}

func (ft *FakerTree) Fake(r *generator.Request) (interface{}, error) {
	return ft.Tree.Fake(r)
}

func (ft *FakerTree) Append(node common.FakerNode) {
	t := &generator.Tree{
		Name:   node.Name(),
		Test:   node.Test,
		Fake:   node.Fake,
		Custom: true,
	}
	ft.Tree.Append(t)
}

func (ft *FakerTree) Insert(index int, node common.FakerNode) error {
	return ft.Tree.Insert(index, &generator.Tree{
		Name:   node.Name(),
		Test:   node.Test,
		Fake:   node.Fake,
		Custom: true,
	})
}

func (ft *FakerTree) RemoveAt(index int) error {
	return ft.Tree.RemoveAt(index)
}

func (ft *FakerTree) Remove(name string) error {
	return ft.Tree.Remove(name)
}
