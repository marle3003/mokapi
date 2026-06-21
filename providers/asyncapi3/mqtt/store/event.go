package store

type Event struct {
	Api    string
	Topic  string
	Retain bool
	Value  string
}
