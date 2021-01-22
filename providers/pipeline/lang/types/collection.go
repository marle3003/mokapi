package types

type Collection interface {
	Children() *Array
	//depthFirst() *Array
}
