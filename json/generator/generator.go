package generator

import (
	"fmt"
	"math/rand"
	"time"
)

var g = &generator{
	rand: rand.New(rand.NewSource(time.Now().Unix())),
	tree: NewTree(),
}

type generator struct {
	rand *rand.Rand

	tree *Tree
}

func Seed(seed int64) {
	g.rand.Seed(seed)
}

func New(r *Request) (interface{}, error) {
	r.g = g
	r.context = map[string]interface{}{}
	return r.g.tree.Resolve(r)
}

func FindByName(name string) *Tree {
	if len(name) == 0 {
		return g.tree
	}
	if g.tree.Name == name {
		return g.tree
	}
	return g.tree.FindByName(name)
}

func (t *Tree) FindByName(name string) *Tree {
	for _, node := range t.Nodes {
		if node.Name == name {
			return node
		}
		if n := node.FindByName(name); n != nil {
			return n
		}
	}
	return nil
}

func (t *Tree) Append(node *Tree) {
	t.Nodes = append(t.Nodes, node)
}

func (t *Tree) Insert(index int, node *Tree) error {
	if index < 0 {
		return fmt.Errorf("index must be positive: %v", index)
	}
	if index >= len(t.Nodes) {
		return fmt.Errorf("index outside of array: %v", index)
	}
	t.Nodes = append(t.Nodes[:index+1], t.Nodes[index:]...)
	t.Nodes[index] = node
	return nil
}

func (t *Tree) Remove(index int) error {
	if index < 0 {
		return fmt.Errorf("index must be positive: %v", index)
	}
	if index >= len(t.Nodes) {
		return fmt.Errorf("index outside of array: %v", index)
	}
	t.Nodes = append(t.Nodes[:index], t.Nodes[index+1:]...)
	return nil
}
