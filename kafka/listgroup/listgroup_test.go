package listgroup_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"testing"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.ListGroup]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(4), reg.MaxVersion)
}
