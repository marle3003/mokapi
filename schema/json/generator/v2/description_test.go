package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestStringDescription(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "description",
			req: &Request{
				Path:   []string{"description"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ourselves whomever wade regularly you how theirs these tomorrow staff gloves wow then opposite conclude those abroad she stop mob a rubbish mob as.", v)
			},
		},
		{
			name: "description with max length",
			req: &Request{
				Path:   []string{"description"},
				Schema: schematest.New("string", schematest.WithMaxLength(50)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				s := v.(string)
				require.Less(t, len(s), 51)
				require.Equal(t, "Say just these run whose foot this least.", v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
