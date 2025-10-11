package fetch_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"testing"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Fetch]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(12), reg.MaxVersion)
}
