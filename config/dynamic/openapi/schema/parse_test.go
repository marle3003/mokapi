package schema

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestParseString(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"int",
			func(t *testing.T) {
				i, err := ParseString("42", &Ref{Value: &Schema{Type: "integer"}})
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			"int64",
			func(t *testing.T) {
				i, err := ParseString("42", &Ref{Value: &Schema{Type: "integer", Format: "int64"}})
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			"int32",
			func(t *testing.T) {
				i, err := ParseString("42", &Ref{Value: &Schema{Type: "integer", Format: "int32"}})
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			"int32 max overflow",
			func(t *testing.T) {
				n := int64(math.MaxInt32) + 1
				_, err := ParseString(fmt.Sprintf("%v", n), &Ref{Value: &Schema{Type: "integer", Format: "int32"}})
				require.EqualError(t, err, "could not parse '2147483648', represents a number either less than int32 min value or greater max value, expected schema type=integer format=int32")
			},
		},
		{
			"int32 min overflow",
			func(t *testing.T) {
				n := int64(math.MinInt32) - 1
				_, err := ParseString(fmt.Sprintf("%v", n), &Ref{Value: &Schema{Type: "integer", Format: "int32"}})
				require.EqualError(t, err, "could not parse '-2147483649', represents a number either less than int32 min value or greater max value, expected schema type=integer format=int32")
			},
		},
		{
			"int but float",
			func(t *testing.T) {
				_, err := ParseString("3.141", &Ref{Value: &Schema{Type: "integer"}})
				require.EqualError(t, err, "could not parse '3.141' as int, expected schema type=integer")
			},
		},
		{
			"int32 but float",
			func(t *testing.T) {
				_, err := ParseString("3.141", &Ref{Value: &Schema{Type: "integer", Format: "int32"}})
				require.EqualError(t, err, "could not parse '3.141' as int, expected schema type=integer format=int32")
			},
		},
		{
			"float default",
			func(t *testing.T) {
				i, err := ParseString("3.141", &Ref{Value: &Schema{Type: "number"}})
				require.NoError(t, err)
				require.Equal(t, 3.141, i)
			},
		},
		{
			"double",
			func(t *testing.T) {
				i, err := ParseString("3.141", &Ref{Value: &Schema{Type: "number", Format: "double"}})
				require.NoError(t, err)
				require.Equal(t, 3.141, i)
			},
		},
		{
			"float",
			func(t *testing.T) {
				i, err := ParseString("3.141", &Ref{Value: &Schema{Type: "number", Format: "float"}})
				require.NoError(t, err)
				require.Equal(t, 3.141, i)
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}
