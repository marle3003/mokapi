package parser

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestError(t *testing.T) {
	testcases := []struct {
		name string
		err  Error
		exp  string
	}{
		{
			name: "one error",
			err: Error{
				err: &ErrorList{
					fmt.Errorf("string 'a' is less than minimum of 3"),
				},
			},
			exp: "error count 1:\n- string 'a' is less than minimum of 3",
		},
		{
			name: "two error",
			err: Error{
				err: &ErrorList{
					fmt.Errorf("string 'a' is less than minimum of 3"),
					fmt.Errorf("item count 2 exceeds maximum count of 1"),
				},
			},
			exp: "error count 2:\n- string 'a' is less than minimum of 3\n- item count 2 exceeds maximum count of 1",
		},
		{
			name: "error detail",
			err: Error{
				err: &ErrorList{
					&ErrorDetail{
						Message: "",
						Field:   "foo",
						Errors:  ErrorList{fmt.Errorf("string 'a' is less than minimum of 3")},
					},
				},
			},
			exp: "error count 1:\n- #/foo:\n\t- string 'a' is less than minimum of 3",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.err.Error()
			require.Equal(t, tc.exp, s)
		})
	}
}
