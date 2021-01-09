package types

type Collection interface {
	Find(match Predicate) (Object, error)
	FindAll(match Predicate) (*Array, error)
	Children() *Array
	//depthFirst() *Array
}
