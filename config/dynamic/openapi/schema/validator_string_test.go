package schema_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
	"testing"
)

func TestParse_String(t *testing.T) {
	maxLength2 := 2
	maxLength3 := 3

	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		err    error
	}{
		{
			"not string",
			`12`,
			&schema.Schema{Type: "string"},
			fmt.Errorf("could not parse 12 as string, expected schema type=string"),
		},
		{
			"string",
			`"gbRMaRxHkiJBPta"`,
			&schema.Schema{Type: "string"},
			nil,
		},
		{
			"type not defined",
			`"gbRMaRxHkiJBPta"`,
			&schema.Schema{},
			nil,
		},
		{
			"by pattern",
			`"013-64-5994"`,
			&schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			nil,
		},
		{
			"not pattern",
			`"013-64-59943"`,
			&schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			fmt.Errorf("value '013-64-59943' does not match pattern, expected schema type=string pattern=^\\d{3}-\\d{2}-\\d{4}$"),
		},
		{
			"date",
			`"1908-12-07"`,
			&schema.Schema{Type: "string", Format: "date"},
			nil,
		},
		{
			"not date",
			`"1908-12-7"`,
			&schema.Schema{Type: "string", Format: "date"},
			fmt.Errorf("value '1908-12-7' is not a date RFC3339, expected schema type=string format=date"),
		},
		{
			"date-time",
			`"1908-12-07T04:14:25Z"`,
			&schema.Schema{Type: "string", Format: "date-time"},
			nil,
		},
		{
			"not date-time",
			`"1908-12-07 T04:14:25Z"`,
			&schema.Schema{Type: "string", Format: "date-time"},
			fmt.Errorf("value '1908-12-07 T04:14:25Z' is not a date-time RFC3339, expected schema type=string format=date-time"),
		},
		{
			"password",
			`"H|$9lb{J<+S;"`,
			&schema.Schema{Type: "string", Format: "password"},
			nil,
		},
		{
			"email",
			`"markusmoen@pagac.net"`,
			&schema.Schema{Type: "string", Format: "email"},
			nil,
		},
		{
			"not email",
			`"markusmoen@@pagac.net"`,
			&schema.Schema{Type: "string", Format: "email"},
			fmt.Errorf("value 'markusmoen@@pagac.net' is not an email address, expected schema type=string format=email"),
		},
		{
			"uuid",
			`"590c1440-9888-45b0-bd51-a817ee07c3f2"`,
			&schema.Schema{Type: "string", Format: "uuid"},
			nil,
		},
		{
			"not uuid",
			`"590c1440-9888-45b0-bd51-a817ee07c3f2a"`,
			&schema.Schema{Type: "string", Format: "uuid"},
			fmt.Errorf("value '590c1440-9888-45b0-bd51-a817ee07c3f2a' is not an uuid, expected schema type=string format=uuid"),
		},
		{
			"ipv4",
			`"152.23.53.100"`,
			&schema.Schema{Type: "string", Format: "ipv4"},
			nil,
		},
		{
			"not ipv4",
			`"152.23.53.100."`,
			&schema.Schema{Type: "string", Format: "ipv4"},
			fmt.Errorf("value '152.23.53.100.' is not an ipv4, expected schema type=string format=ipv4"),
		},
		{
			"ipv6",
			`"8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			&schema.Schema{Type: "string", Format: "ipv6"},
			nil,
		},
		{
			"not ipv6",
			`"-8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			&schema.Schema{Type: "string", Format: "ipv6"},
			fmt.Errorf("value '-8898:ee17:bc35:9064:5866:d019:3b95:7857' is not an ipv6, expected schema type=string format=ipv6"),
		},
		{
			"not minLength",
			`"foo"`,
			&schema.Schema{Type: "string", MinLength: toIntP(4)},
			fmt.Errorf("value 'foo' does not meet min length of 4"),
		},
		{
			"minLength",
			`"foo"`,
			&schema.Schema{Type: "string", MinLength: toIntP(3)},
			nil,
		},
		{
			"not maxLength",
			`"foo"`,
			&schema.Schema{Type: "string", MaxLength: &maxLength2},
			fmt.Errorf("value 'foo' does not meet max length of 2"),
		},
		{
			"maxLength",
			`"foo"`,
			&schema.Schema{Type: "string", MaxLength: &maxLength3},
			nil,
		},
		{
			"enum",
			`"foo"`,
			&schema.Schema{Type: "string", Enum: []interface{}{"foo"}},
			nil,
		},
		{
			"not in enum",
			`"foo"`,
			&schema.Schema{Type: "string", Enum: []interface{}{"bar"}},
			fmt.Errorf("value foo does not match one in the enum [bar]"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.Equal(t, d.err, err)
			require.Equal(t, d.s[1:len(d.s)-1], i)
		})
	}
}
