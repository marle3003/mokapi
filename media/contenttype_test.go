package media

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseContentType(t *testing.T) {
	testcases := []struct {
		s        string
		validate func(t *testing.T, ct ContentType)
	}{
		{
			s: "",
			validate: func(t *testing.T, ct ContentType) {
				require.Equal(t, "", ct.Type)
				require.Equal(t, "", ct.Subtype)
				require.Equal(t, "", ct.Key())
				require.True(t, ct.IsRange())
			},
		},
		{
			s: "text/plain",
			validate: func(t *testing.T, ct ContentType) {
				require.Equal(t, "text", ct.Type)
				require.Equal(t, "plain", ct.Subtype)
				require.Equal(t, "text/plain", ct.Key())
				require.False(t, ct.IsRange())
			},
		},
		{
			s: "*/*",
			validate: func(t *testing.T, ct ContentType) {
				require.Equal(t, "*", ct.Type)
				require.Equal(t, "*", ct.Subtype)
				require.Equal(t, "*/*", ct.Key())
				require.True(t, ct.IsRange())
			},
		},
		{
			s: "text/*",
			validate: func(t *testing.T, ct ContentType) {
				require.Equal(t, "text", ct.Type)
				require.Equal(t, "*", ct.Subtype)
				require.Equal(t, "text/*", ct.Key())
				require.True(t, ct.IsRange())
			},
		},
		{
			s: "application/xhtml+xml",
			validate: func(t *testing.T, ct ContentType) {
				require.Equal(t, "application", ct.Type)
				require.Equal(t, "xhtml+xml", ct.Subtype)
				require.Equal(t, "application/xhtml+xml", ct.Key())
				require.False(t, ct.IsRange())
			},
		},
		{
			s: "application/xml;q=0.9",
			validate: func(t *testing.T, ct ContentType) {
				require.Equal(t, "application", ct.Type)
				require.Equal(t, "xml", ct.Subtype)
				require.Equal(t, 0.9, ct.Q)
				require.Equal(t, "application/xml", ct.Key())
				require.Equal(t, "application/xml;q=0.9", ct.String())
				require.False(t, ct.IsRange())
			},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.s, func(t *testing.T) {
			ct := ParseContentType(test.s)
			test.validate(t, ct)
		})
	}
}

func TestContentType_Match(t *testing.T) {
	testcases := []struct {
		name string
		a    string
		b    string
		exp  bool
	}{
		{
			"text/plain - text/plain",
			"text/plain",
			"text/plain",
			true,
		},
		{
			"text/plain - */*",
			"text/plain",
			"*/*",
			true,
		},
		{
			"*/* - text/plain",
			"*/*",
			"text/plain",
			true,
		},
		{
			"text/plain - text/*",
			"text/plain",
			"text/*",
			true,
		},
		{
			"text/* - text/plain",
			"text/*",
			"text/plain",
			true,
		},
		{
			"text/plain;format=flowed - text/plain",
			"text/plain;format=flowed",
			"text/plain",
			true,
		},
		{
			"text/plain - text/plain; format=flowed",
			"text/plain",
			"text/plain; format=flowed",
			true,
		},
		{
			"text/plain;format=flowed - text/plain;format=flowed",
			"text/plain;format=flowed",
			"text/plain;format=flowed",
			true,
		},
		{
			"text/plain;format=flowed - text/plain;format=fixed",
			"text/plain;format=flowed",
			"text/plain;format=fixed",
			false,
		},
		{
			"text/plain - image/png",
			"text/plain",
			"image/png",
			false,
		},
		{
			"text/plain - text/html",
			"text/plain",
			"text/html",
			false,
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			m := ParseContentType(test.a).Match(ParseContentType(test.b))
			require.Equal(t, test.exp, m)
		})

	}
}
