package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestStringFormat(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "date",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("date")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1977-05-07", v)
			},
		},
		{
			name: "date-time",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("date-time")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1977-05-07T20:13:28Z", v)
			},
		},
		{
			name: "time",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("time")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "20:13:28Z", v)
			},
		},
		{
			name: "password",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("password")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "d*SzbmJ2YPhc", v)
			},
		},
		{
			name: "email",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("email")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "annalise.hermann@elliott.org", v)
			},
		},
		{
			name: "uuid",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("uuid")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "ce702a60-0f08-4819-bcd4-0907c044ad5c", v)
			},
		},
		{
			name: "{url}",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("{url}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "uri",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("uri")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "hostname",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("hostname")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "central24-7.biz", v)
			},
		},
		{
			name: "ipv4",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("ipv4")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "206.188.64.71", v)
			},
		},
		{
			name: "ipv6",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("ipv6")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "70ce:d4bc:cf40:8e47:6e2d:abe:70be:4dc9", v)
			},
		},
		{
			name: "{beername}",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("{beername}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Stone Imperial Russian Stout", v)
			},
		},
		{
			name: "{zip} {city}",
			req:  &Request{Schema: schematest.New("string", schematest.WithFormat("{zip} {city}"))},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "32824 San Jose", v)
			},
		},
		{
			name: "string enum",
			req: &Request{
				Schema: schematest.New("string", schematest.WithEnumValues("foo", "bar")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
