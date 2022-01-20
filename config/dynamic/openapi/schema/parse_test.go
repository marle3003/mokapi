package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseString_Number(t *testing.T) {
	i, err := ParseString("3.612", &Ref{Value: &Schema{Type: "number"}})
	require.NoError(t, err)
	require.Equal(t, 3.612, i)
}
