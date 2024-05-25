package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schematest"
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
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("date"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2001-08-26", v)
			},
		},
		{
			name: "date-time",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("date-time"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2001-08-26T07:02:11Z", v)
			},
		},
		{
			name: "password",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("password"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "w*YoR94jFL@X", v)
			},
		},
		{
			name: "email",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("email"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "shanellewehner@cruickshank.biz", v)
			},
		},
		{
			name: "uuid",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("uuid"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "{url}",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("{url}"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "uri",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("uri"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "hostname",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("hostname"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "corporatevirtual.com", v)
			},
		},
		{
			name: "ipv4",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("ipv4"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "68.244.174.93", v)
			},
		},
		{
			name: "ipv6",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("ipv6"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1944:e0f4:7bae:825d:7d23:7d3e:448f:6589", v)
			},
		},
		{
			name: "{beername}",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("{beername}"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Schneider Aventinus", v)
			},
		},
		{
			name: "{zip} {city}",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewRef("string", schematest.WithFormat("{zip} {city}"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "82910 San Jose", v)
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

func TestName(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "name",
			req:  &Request{Path: Path{&PathElement{Name: "name"}}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "name as string",
			req: &Request{Path: Path{
				&PathElement{Name: "name", Schema: schematest.NewRef("string")},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "customerName",
			req:  &Request{Path: Path{&PathElement{Name: "customerName"}}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "campaignName",
			req: &Request{Path: Path{
				&PathElement{Name: "campaignName", Schema: schematest.NewRef("string")},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "name with max length 3",
			req: &Request{Path: Path{
				&PathElement{
					Name: "name",
					Schema: schematest.NewRef("string",
						schematest.WithMaxLength(3),
					),
				},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Neo", v)
			},
		},
		{
			name: "name with min and max length 4",
			req: &Request{Path: Path{
				&PathElement{
					Name: "name",
					Schema: schematest.NewRef("string",
						schematest.WithMinLength(4),
						schematest.WithMaxLength(4),
					),
				},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Apex", v)
			},
		},
		{
			name: "name with min length 4",
			req: &Request{Path: Path{
				&PathElement{
					Name: "name",
					Schema: schematest.NewRef("string",
						schematest.WithMinLength(4),
					),
				},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "TerraCove", v)
			},
		},
		{
			name: "name with min and max length 5",
			req: &Request{Path: Path{
				&PathElement{
					Name: "name",
					Schema: schematest.NewRef("string",
						schematest.WithMinLength(5),
						schematest.WithMaxLength(5),
					),
				},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Exalt", v)
			},
		},
		{
			name: "name with min and max length 6",
			req: &Request{Path: Path{
				&PathElement{
					Name: "name",
					Schema: schematest.NewRef("string",
						schematest.WithMinLength(6),
						schematest.WithMaxLength(6),
					),
				},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Spirit", v)
			},
		},
		{
			name: "name with min length 7",
			req: &Request{Path: Path{
				&PathElement{
					Name: "name",
					Schema: schematest.NewRef("string",
						schematest.WithMinLength(7),
					),
				},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "EchoValley", v)
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

func TestStringId(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "id",
			req: &Request{
				Path: Path{
					&PathElement{Name: "id", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "id string with max",
			req: &Request{
				Path: Path{
					&PathElement{
						Name:   "id",
						Schema: schematest.NewRef("string", schematest.WithMaxLength(30)),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "802910936489301180573501861460", v)
			},
		},
		{
			name: "id string with min",
			req: &Request{
				Path: Path{
					&PathElement{
						Name:   "id",
						Schema: schematest.NewRef("string", schematest.WithMinLength(4)),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "id string with min & max",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "id",
						Schema: schematest.NewRef("string",
							schematest.WithMinLength(4),
							schematest.WithMaxLength(10),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7291093", v)
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

func TestStringNumber(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "partyNumber",
			req: &Request{
				Path: Path{
					&PathElement{Name: "partyNumber", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "80291093648", v)
			},
		},
		{
			name: "partyNumbers as array",
			req: &Request{
				Path: Path{
					&PathElement{
						Name:   "partyNumbers",
						Schema: schematest.NewRef("array", schematest.WithItems("string"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"22910936489", "71180573501"}, v)
			},
		},
		{
			name: "employeeNumber with min=max",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "id",
						Schema: schematest.NewRef("string",
							schematest.WithMinLength(8),
							schematest.WithMaxLength(8),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, v, 8)
				require.Equal(t, "80291093", v)
			},
		},
		{
			name: "id string with min",
			req: &Request{
				Path: Path{
					&PathElement{
						Name:   "id",
						Schema: schematest.NewRef("string", schematest.WithMinLength(4)),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "id string with min & max",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "id",
						Schema: schematest.NewRef("string",
							schematest.WithMinLength(4),
							schematest.WithMaxLength(10),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7291093", v)
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

func TestStringKey(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "businessKey",
			req: &Request{
				Path: Path{
					&PathElement{Name: "businessKey", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "key with pattern",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "businessKey",
						Schema: schematest.NewRef("string",
							schematest.WithPattern("[a-z]{3}"),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "cqo", v)
			},
		},
		{
			name: "key string with min",
			req: &Request{
				Path: Path{
					&PathElement{
						Name:   "key",
						Schema: schematest.NewRef("string", schematest.WithMinLength(4)),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "key string with min & max",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "key",
						Schema: schematest.NewRef("string",
							schematest.WithMinLength(4),
							schematest.WithMaxLength(10),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7291093", v)
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

func TestStringEmail(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "email",
			req: &Request{
				Path: Path{
					&PathElement{Name: "email", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "shanellewehner@cruickshank.biz", v)
			},
		},
		{
			name: "email no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "email"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "shanellewehner@cruickshank.biz", v)
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

func TestStringUrl(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "uri",
			req: &Request{
				Path: Path{
					&PathElement{Name: "uri", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "uri no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "uri"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "url",
			req: &Request{
				Path: Path{
					&PathElement{Name: "uri", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "url no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "uri"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "curl no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "curl"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3652171958352792229), v)
			},
		},
		{
			name: "updateUrl",
			req: &Request{
				Path: Path{
					&PathElement{Name: "curl", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "updateURL",
			req: &Request{
				Path: Path{
					&PathElement{Name: "curl", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "url",
			req: &Request{
				Path: Path{
					&PathElement{Name: "url", Schema: schematest.NewRef("array", schematest.WithItems("string"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
			},
		},
		{
			name: "urls",
			req: &Request{
				Path: Path{
					&PathElement{Name: "url", Schema: schematest.NewRef("array", schematest.WithItems("string"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
			},
		},
		{
			name: "photoUrls",
			req: &Request{
				Path: Path{
					&PathElement{Name: "url", Schema: schematest.NewRef("array", schematest.WithItems("string"))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
			},
		},
		{
			name: "urls no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "urls"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
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

func TestStringError(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "error",
			req: &Request{
				Path: Path{
					&PathElement{Name: "error", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "parameter not initialized", v)
			},
		},
		{
			name: "error no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "error"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "parameter not initialized", v)
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
				Path: Path{
					&PathElement{Name: "hash", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "54686520636f6d70616e7920736d656c6c2eda39a3ee5e6b4b0d3255bfef95601890afd80709", v)
			},
		},
		{
			name: "error no schema",
			req: &Request{
				Path: Path{
					&PathElement{Name: "error"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "parameter not initialized", v)
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

func TestStringDescription(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "description",
			req: &Request{
				Path: Path{
					&PathElement{Name: "description", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Tonight daily regularly you for recline her choir insufficient poised tribe Caesarian mustering will might.", v)
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
