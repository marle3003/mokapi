package store

import (
	"mokapi/test"
	"testing"
)

func TestGroup(t *testing.T) {
	g := &Group{
		name: "foo",
	}
	test.Equals(t, "foo", g.Name())
	test.Equals(t, Stable, g.State())

	g.NewGeneration()
	test.Assert(t, g.Generation() != nil, "generation not nil")
	test.Equals(t, 0, g.Generation().Id)

	g.NewGeneration()
	test.Assert(t, g.Generation() != nil, "generation not nil")
	test.Equals(t, 1, g.Generation().Id)

	g.SetState(Joining)
	test.Equals(t, Joining, g.State())
}

func TestGroup_Commit(t *testing.T) {
	g := &Group{}
	test.Equals(t, int64(-1), g.Offset("foo", 0))
	test.Equals(t, int64(-1), g.Offset("foo", 1))

	g.Commit("foo", 0, 1)
	g.Commit("foo", 1, 10)

	test.Equals(t, int64(1), g.Offset("foo", 0))
	test.Equals(t, int64(10), g.Offset("foo", 1))
}
