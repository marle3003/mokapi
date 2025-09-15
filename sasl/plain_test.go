package sasl_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/sasl"
	"testing"
)

func TestPlain(t *testing.T) {
	c := sasl.NewPlainClient("id", "username", "password")
	s := sasl.NewPlainServer(func(identity, username, password string) error {
		require.Equal(t, "id", identity)
		require.Equal(t, "username", username)
		require.Equal(t, "password", password)
		return nil
	})
	b, err := c.Next(nil)
	require.NoError(t, err)
	r, err := s.Next(b)
	require.NoError(t, err)
	require.Nil(t, r)

}
