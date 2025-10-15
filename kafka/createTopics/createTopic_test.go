package createTopics_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"testing"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.CreateTopics]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(7), reg.MaxVersion)
}
