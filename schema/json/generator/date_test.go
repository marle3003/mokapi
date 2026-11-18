package generator

import (
	"mokapi/schema/json/schema/schematest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestStringDate(t *testing.T) {
	// tests depends on current year so without this, all tests will break in next year
	isDateString := func(t *testing.T, s any) {
		str, ok := s.(string)
		require.True(t, ok, "expected string")
		_, err := time.Parse("2006-01-02", str)
		require.NoError(t, err)
	}
	isDateTimeString := func(t *testing.T, s any) {
		_, err := time.Parse(time.RFC3339, s.(string))
		require.NoError(t, err)
	}

	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "created",
			req: &Request{
				Path:   []string{"created"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "createdAt",
			req: &Request{
				Path: []string{"createdAt"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "creationDate",
			req: &Request{
				Path: []string{"creationDate"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "modified",
			req: &Request{
				Path: []string{"modified"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "modifyDate",
			req: &Request{
				Path: []string{"modifyDate"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "updateDate",
			req: &Request{
				Path: []string{"updateDate"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "deleted",
			req: &Request{
				Path: []string{"deleted"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "deletedAt",
			req: &Request{
				Path: []string{"deletedAt"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "deleteDate",
			req: &Request{
				Path: []string{"deleteDate"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "foundationDate",
			req: &Request{
				Path: []string{"foundationDate"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "activeFrom and inactiveFrom",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("activeFrom", schematest.New("string", schematest.WithFormat("date-time"))),
					schematest.WithProperty("inactiveFrom", schematest.New("string", schematest.WithFormat("date-time"))),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				m := v.(map[string]any)
				isDateTimeString(t, m["activeFrom"])
				isDateTimeString(t, m["inactiveFrom"])
			},
		},
		{
			name: "publishedFrom and publishedUntil",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("publishedFrom", schematest.New("string", schematest.WithFormat("date-time"))),
					schematest.WithProperty("publishedUntil", schematest.New("string", schematest.WithFormat("date-time"))),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				m := v.(map[string]any)
				isDateTimeString(t, m["publishedFrom"])
				isDateTimeString(t, m["publishedUntil"])
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

func TestFakeDateWithYearRange_NewYearsEve(t *testing.T) {
	r, err := fakeDateWithYearRange(&Request{}, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), 2027)
	require.NoError(t, err)
	require.NotEqual(t, "", r)
	require.True(t, strings.HasPrefix(r.(string), "2027"))
}

func TestTimePairs(t *testing.T) {
	pairs := []struct{ start, end string }{
		{"startDate", "endDate"},
		{"validFrom", "validUntil"},
	}
	formats := map[string]string{
		"":          "2006-01-02",
		"date":      "2006-01-02",
		"date-time": time.RFC3339,
	}

	for _, pair := range pairs {
		for format, layout := range formats {
			for _, order := range [][]string{{pair.start, pair.end}, {pair.end, pair.start}} {
				name := pair.start + "/" + pair.end + " format=" + format + " order=" + strings.Join(order, ",")
				t.Run(name, func(t *testing.T) {
					Seed(1234567)

					s := schematest.New("object", schematest.WithRequired(order...))
					for _, prop := range order {
						opts := []schematest.SchemaOptions{}
						if format != "" {
							opts = append(opts, schematest.WithFormat(format))
						}
						schematest.WithProperty(prop, schematest.New("string", opts...))(s)
					}

					v, err := New(&Request{Schema: s})

					require.NoError(t, err)
					obj := v.(map[string]any)
					start, err := time.Parse(layout, obj[pair.start].(string))
					require.NoError(t, err, "%v", obj)
					end, err := time.Parse(layout, obj[pair.end].(string))
					require.NoError(t, err, "%v", obj)
					require.False(t, end.Before(start), "%v", obj)
				})
			}
		}
	}
}

func TestTimePairs_EndDerivedFromStart(t *testing.T) {
	for _, format := range []string{"", "date", "date-time"} {
		t.Run("format="+format, func(t *testing.T) {
			Seed(1234567)

			opts := []schematest.SchemaOptions{}
			if format != "" {
				opts = append(opts, schematest.WithFormat(format))
			}
			v, err := New(&Request{Schema: schematest.New("array",
				schematest.WithItems("object", schematest.WithProperty("endDate", schematest.New("string", opts...)),
					schematest.WithProperty("startDate", schematest.New("string", opts...)),
					schematest.WithRequired("endDate", "startDate")),
				schematest.WithMinItems(40), schematest.WithMaxItems(40),
			)})

			require.NoError(t, err)
			layout := "2006-01-02"
			if format == "date-time" {
				layout = time.RFC3339
			}
			parse := func(s string) (time.Time, error) { return time.Parse(layout, s) }
			endsInThePast := 0
			for _, e := range v.([]any) {
				obj := e.(map[string]any)
				start, err := parse(obj["startDate"].(string))
				require.NoError(t, err, "%v", obj)
				end, err := parse(obj["endDate"].(string))
				require.NoError(t, err, "%v", obj)
				require.False(t, end.Before(start), "%v", obj)
				if end.Before(time.Now()) {
					endsInThePast++
				}
			}
			// an end that ignores its start would always lie in the future
			require.Greater(t, endsInThePast, 0)
		})
	}
}
