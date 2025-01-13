package directory

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAttributeType(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		test  func(t *testing.T, attrType AttributeType, err error)
	}{
		{
			name:  "cn",
			input: "( 2.5.4.3 NAME 'cn' DESC 'Common Name' EQUALITY caseIgnoreMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{64} )",
			test: func(t *testing.T, attrType AttributeType, err error) {
				require.NoError(t, err)
				require.Equal(t, "2.5.4.3", attrType.Id)
				require.Equal(t, []string{"cn"}, attrType.Name)
				require.Equal(t, "Common Name", attrType.Description)
				require.Equal(t, "caseIgnoreMatch", attrType.Equality)
				require.Equal(t, "1.3.6.1.4.1.1466.115.121.1.15{64}", attrType.Syntax)
			},
		},
		{
			name:  "multiple names",
			input: "( 2.5.4.4 NAME ( 'sn' 'surname' ) DESC 'Surname' EQUALITY caseIgnoreMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{64} )",
			test: func(t *testing.T, attrType AttributeType, err error) {
				require.NoError(t, err)
				require.Equal(t, "2.5.4.4", attrType.Id)
				require.Equal(t, []string{"sn", "surname"}, attrType.Name)
				require.Equal(t, "Surname", attrType.Description)
				require.Equal(t, "caseIgnoreMatch", attrType.Equality)
				require.Equal(t, "1.3.6.1.4.1.1466.115.121.1.15{64}", attrType.Syntax)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			attrType, err := parseAttributeType(tc.input)
			tc.test(t, attrType, err)
		})
	}
}
