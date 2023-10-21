package xml

import "mokapi/sortedmap"

type Node struct {
	Name       string
	Children   []*Node
	Attributes *sortedmap.LinkedHashMap[string, string]
	Content    string
}

func NewNode(name string) *Node {
	return &Node{
		Name:       name,
		Attributes: &sortedmap.LinkedHashMap[string, string]{},
	}
}

func (n *Node) GetFirstElement(name string) *Node {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}
