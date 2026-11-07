package store

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"sync"
	"testing"
)

func TestGroup_NewGeneration(t *testing.T) {
	logrus.SetOutput(io.Discard)
	g := &Group{}

	generation := g.NewGeneration()
	require.Equal(t, 0, generation.Id)
	require.Nil(t, g.Generation(), "a new generation is not published before SetGeneration")

	g.SetGeneration(generation)
	require.Same(t, generation, g.Generation())

	require.Equal(t, 1, g.NewGeneration().Id)
}

func TestGroup_AssignPartitions(t *testing.T) {
	g := &Group{}
	generation := g.NewGeneration()
	generation.Members["m1"] = newMember(nil, 1000)
	g.SetGeneration(generation)

	g.AssignPartitions("m1", map[string][]int{"foo": {1, 2}})

	require.Equal(t, map[string][]int{"foo": {1, 2}}, g.Snapshot().Generation.Members["m1"].Partitions)
}

func TestGroup_Snapshot(t *testing.T) {
	logrus.SetOutput(io.Discard)
	g := &Group{Name: "foo"}
	generation := g.NewGeneration()
	generation.Members["m1"] = newMember(nil, 1000)
	g.SetGeneration(generation)
	g.SetState(Stable)
	g.AssignPartitions("m1", map[string][]int{"topic": {1}})
	g.Commit("topic", 1, 5)

	snapshot := g.Snapshot()
	require.Equal(t, "foo", snapshot.Name)
	require.Equal(t, GroupState(Stable), snapshot.State)
	require.Equal(t, map[string]map[int]int64{"topic": {1: 5}}, snapshot.Commits)

	// changes to the group must not show up in an earlier snapshot
	g.AssignPartitions("m1", map[string][]int{"topic": {2}, "other": {3}})
	g.Commit("topic", 1, 6)
	g.SetState(Empty)

	require.Equal(t, GroupState(Stable), snapshot.State)
	require.Equal(t, map[string][]int{"topic": {1}}, snapshot.Generation.Members["m1"].Partitions)
	require.Equal(t, int64(5), snapshot.Commits["topic"][1])
}

func TestGroup_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)
	g := &Group{Name: "foo"}
	generation := g.NewGeneration()
	generation.Members["m1"] = newMember(nil, 1000)
	g.SetGeneration(generation)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				g.SetState(GroupState(j % 4))
				g.SetGeneration(generation)
				g.AssignPartitions("m1", map[string][]int{"topic": {j}})
				g.Commit("topic", j%3, int64(j))
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = g.State()
				_ = g.Generation()
				_ = g.Snapshot()
				_ = g.Offset("topic", j%3)
				_ = g.NewGeneration()
			}
		}()
	}
	wg.Wait()
}

func TestGroup_Commit(t *testing.T) {
	logrus.SetOutput(io.Discard)

	g := &Group{}
	require.Equal(t, int64(-1), g.Offset("foo", 0))
	require.Equal(t, int64(-1), g.Offset("foo", 1))

	g.Commit("foo", 0, 1)
	g.Commit("foo", 1, 10)

	require.Equal(t, int64(1), g.Offset("foo", 0))
	require.Equal(t, int64(10), g.Offset("foo", 1))
}
