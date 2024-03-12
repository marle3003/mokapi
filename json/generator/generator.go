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
	return g.tree.FindByName(name)
}

func (t *Tree) FindByName(name string) *Tree {
	for _, node := range t.nodes {
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
	t.nodes = append(t.nodes, node)
}

func (t *Tree) Insert(index int, node *Tree) error {
	if index < 0 {
		return fmt.Errorf("index must be positive: %v", index)
	}
	if index >= len(t.nodes) {
		return fmt.Errorf("index outside of array: %v", index)
	}
	t.nodes = append(t.nodes[:index+1], t.nodes[index:]...)
	t.nodes[index] = node
	return nil
}

func (t *Tree) Remove(index int) error {
	if index < 0 {
		return fmt.Errorf("index must be positive: %v", index)
	}
	if index >= len(t.nodes) {
		return fmt.Errorf("index outside of array: %v", index)
	}
	t.nodes = append(t.nodes[:index], t.nodes[index+1:]...)
	return nil
}
