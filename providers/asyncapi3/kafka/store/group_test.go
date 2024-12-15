package store

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestGroup_NewGeneration(t *testing.T) {
	logrus.SetOutput(io.Discard)
	g := &Group{}

	g.NewGeneration()
	require.NotNil(t, g.Generation, "generation not nil")
	require.Equal(t, 0, g.Generation.Id)

	g.NewGeneration()
	require.NotNil(t, g.Generation, "generation not nil")
	require.Equal(t, 1, g.Generation.Id)
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
