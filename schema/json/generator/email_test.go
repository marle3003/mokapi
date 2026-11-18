package generator

import (
	"mokapi/schema/json/schema/schematest"
	"net/mail"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestStringEmail(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "email",
			req: &Request{
				Path:   []string{"email"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "annalise.hermann@elliott.org", v)
			},
		},
		{
			name: "email no schema",
			req: &Request{
				Path: []string{"email"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "annalise.hermann@elliott.org", v)
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

func TestEmailAddressProperty(t *testing.T) {
	t.Run("uses the person name like email does", func(t *testing.T) {
		v, err := New(&Request{
			Schema: schematest.New("object",
				schematest.WithProperty("firstName", schematest.New("string")),
				schematest.WithProperty("lastName", schematest.New("string")),
				schematest.WithProperty("emailAddress", schematest.New("string")),
				schematest.WithRequired("firstName", "lastName", "emailAddress"),
			),
		})

		require.NoError(t, err)
		obj := v.(map[string]any)
		first := strings.ToLower(obj["firstName"].(string))
		last := strings.ToLower(obj["lastName"].(string))
		require.True(t, strings.HasPrefix(obj["emailAddress"].(string), first+"."+last+"@"), "%v", obj)
	})

	t.Run("array items are emails", func(t *testing.T) {
		v, err := New(&Request{
			Path: []string{"emailAddresses"},
			Schema: schematest.New("array",
				schematest.WithItems("string"),
				schematest.WithMinItems(3),
				schematest.WithMaxItems(3),
			),
		})

		require.NoError(t, err)
		items := v.([]any)
		require.Len(t, items, 3)
		for _, item := range items {
			_, err = mail.ParseAddress(item.(string))
			require.NoError(t, err, "%v", item)
		}
	})
}
