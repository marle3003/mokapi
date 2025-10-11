package apiVersion_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"testing"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.ApiVersions]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(3), reg.MaxVersion)
}

func TestNewApiKeyResponse(t *testing.T) {
	res := apiVersion.NewApiKeyResponse(kafka.ApiVersions, kafka.ApiTypes[kafka.ApiVersions])
	require.Equal(t, kafka.ApiVersions, res.ApiKey)
	require.Equal(t, int16(0), res.MinVersion)
	require.Equal(t, int16(3), res.MaxVersion)
}
