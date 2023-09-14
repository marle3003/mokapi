package openapi_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi"
	"testing"
)

func TestResponse_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "tags",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "200": { "description": "foo" } }`), &res)
				require.NoError(t, err)
				require.Equal(t, 1, res.Len())
				r := res.Get(200)
				require.Equal(t, "foo", r.Value.Description)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
