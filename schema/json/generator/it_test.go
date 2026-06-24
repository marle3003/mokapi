package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestIt(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "error",
			req: &Request{
				Path:   []string{"error"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "error", v)
			},
		},
		{
			name: "error no schema",
			req: &Request{
				Path: []string{"error"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "error", v)
			},
		},
		{
			name: "website",
			req: &Request{
				Path:   []string{"website"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "central24-7.biz", v)
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

func TestStringHash(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "hash",
			req: &Request{
				Path:   []string{"hash"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "50726f74656374207468652068616e6420756e6465722063757465206c6f61642eda39a3ee5e6b4b0d3255bfef95601890afd80709", v)
			},
		},
		{
			name: "error no schema",
			req: &Request{
				Path: []string{"error"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "error", v)
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

func TestUser(t *testing.T) {
	// tests depends on current year so without this, all tests will break in next year
	isDateTimeString := func(t *testing.T, s any) {
		_, err := time.Parse(time.RFC3339, s.(string))
		require.NoError(t, err)
	}

	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "user as string",
			req: &Request{
				Path:   []string{"user"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Deanna.Elliott", v)
			},
		},
		{
			name: "user",
			req: &Request{
				Path: []string{"user"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"email":     "annalise.hermann@salesmorph.name",
						"firstname": "Annalise",
						"gender":    "female",
						"lastname":  "Hermann",
						"username":  "ahermann",
					},
					v)
			},
		},
		{
			name: "lastLogin",
			req: &Request{
				Path:   []string{"lastLogin"},
				Schema: schematest.New("string", schematest.WithFormat("date-time")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateTimeString(t, v)
			},
		},
		{
			name: "password",
			req: &Request{
				Path:   []string{"password"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "zPmhSJc2b*Y", v)
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
