package produce_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"testing"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Produce]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(9), reg.MaxVersion)
}
