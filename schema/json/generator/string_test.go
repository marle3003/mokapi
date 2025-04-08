package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
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
				require.Equal(t, "2007-08-26", v)
			},
		},
		{
			name: "date-time",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("date-time")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2007-08-26T07:02:11Z", v)
			},
		},
		{
			name: "password",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("password")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "w*YoR94jFL@X", v)
			},
		},
		{
			name: "email",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("email")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "shanellewehner@cruickshank.biz", v)
			},
		},
		{
			name: "uuid",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("uuid")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "{url}",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("{url}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "uri",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("uri")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "hostname",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("hostname")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "corporatevirtual.com", v)
			},
		},
		{
			name: "ipv4",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("ipv4")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "68.244.174.93", v)
			},
		},
		{
			name: "ipv6",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("ipv6")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1944:e0f4:7bae:825d:7d23:7d3e:448f:6589", v)
			},
		},
		{
			name: "{beername}",
			req: &Request{
				Schema: schematest.New("string", schematest.WithFormat("{beername}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Schneider Aventinus", v)
			},
		},
		{
			name: "{zip} {city}",
			req:  &Request{Schema: schematest.New("string", schematest.WithFormat("{zip} {city}"))},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "82910 San Jose", v)
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
