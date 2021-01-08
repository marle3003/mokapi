package types

type Collection interface {
	Find(match Predicate) (Object, error)
	FindAll(match Predicate) ([]Object, error)
	Children() *Array
	//depthFirst() *Array
}
