package kafka_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"strings"
	"testing"
)

func TestValidateTopicName(t *testing.T) {
	require.EqualError(t, kafka.ValidateTopicName(""), "topic name can not be empty")
	require.EqualError(t, kafka.ValidateTopicName("."), "topic name can not be .")
	require.EqualError(t, kafka.ValidateTopicName(strings.Repeat("a", 250)), "topic name can not be longer than 249")
	require.EqualError(t, kafka.ValidateTopicName("a$"), "topic name is not valid, valid characters are ASCII alphanumerics, '.', '_', and '-'")
	require.NoError(t, kafka.ValidateTopicName("a"))
}
