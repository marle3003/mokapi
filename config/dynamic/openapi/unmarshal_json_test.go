package openapi

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContent_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		s    string
		fn   func(t *testing.T, c Content)
	}{
		{
			"empty",
			"{}",
			func(t *testing.T, c Content) {
				require.Len(t, c, 0)
			},
		},
		{
			"with ref",
			`{
"application/json": {
  "schema": {
    "$ref": "#/components/schemas/Foo"
  }
}
}`,
			func(t *testing.T, c Content) {
				require.Len(t, c, 1)
				require.Contains(t, c, "application/json")
				ct := c["application/json"].ContentType
				require.Equal(t, "application", ct.Type)
				require.Equal(t, "json", ct.Subtype)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := Content{}
			err := json.Unmarshal([]byte(tc.s), &c)
			require.NoError(t, err)
			tc.fn(t, c)
		})
	}
}

func TestResponses_UnmarshalJSON(t *testing.T) {
	s := `{"200": {
"description": "Success"
}}`
	r := &Responses{}
	err := json.Unmarshal([]byte(s), &r)
	require.NoError(t, err)
	require.Equal(t, "Success", r.GetResponse(200).Description)
}
