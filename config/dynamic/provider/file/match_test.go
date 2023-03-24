package file

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMatch(t *testing.T) {
	testcases := []struct {
		pattern string
		test    map[string]bool
	}{
		{
			pattern: "name",
			test: map[string]bool{
				"/name.log":      false,
				"/name/file.txt": true,
				"/lib/name":      true,
				"/lib/name.log":  false,
				"/lib/foo.log":   false,
			},
		},
		{
			pattern: "name/",
			test: map[string]bool{
				"/name/file.txt":     true,
				"/name/log/name.log": true,
				"/name.log":          false,
			},
		},
		{
			pattern: "name.file",
			test: map[string]bool{
				"/name.file":     true,
				"/lib/name.file": true,
				"/name.log":      false,
			},
		},
		{
			pattern: "/name.file",
			test: map[string]bool{
				"/name.file":     true,
				"name.file":      true,
				"/lib/name.file": false,
			},
		},
		{
			pattern: "lib/name.file",
			test: map[string]bool{
				"/lib/name.file":      true,
				"name.file":           false,
				"/test/lib/name.file": false,
			},
		},
		{
			pattern: "**/lib/name.file",
			test: map[string]bool{
				"/lib/name.file":      true,
				"/test/lib/name.file": true,
				"name.file":           false,
			},
		},
		{
			pattern: "**/name",
			test: map[string]bool{
				"/name/log.file":     true,
				"/lib/name/log.file": true,
				"/name/lib/log.file": true,
			},
		},
		{
			pattern: "/lib/**/name",
			test: map[string]bool{
				"/lib/test/ver1/name/log.file": true,
				"/lib/name/log.file":           true,
				"/name/lib/log.file":           false,
			},
		},
		{
			pattern: "*.file",
			test: map[string]bool{
				"/name.file":     true,
				"name.file":      true,
				"/lib/name.file": true,
				"name.txt":       false,
			},
		},
		{
			pattern: "*name/",
			test: map[string]bool{
				"/lastname/log.file":  true,
				"/firstname/log.file": true,
				"/lib/name.file":      false,
				"name.txt":            false,
			},
		},
		{
			// star with slash requires a folder root folder "name" does not match
			pattern: "*/name/",
			test: map[string]bool{
				"/foo/name/log.file": true,
				"/name/log.file":     false,
				"/foo/lib/name.file": false,
				"name.txt":           false,
			},
		},
		{
			pattern: "name?.file",
			test: map[string]bool{
				"/names.file":  true,
				"/name1.file":  true,
				"/names1.file": false,
			},
		},
		{
			pattern: "name[a-z].file",
			test: map[string]bool{
				"/names.file":  true,
				"/nameb.file":  true,
				"/name1.file":  false,
				"/names1.file": false,
			},
		},
		{
			pattern: "name[abc].file",
			test: map[string]bool{
				"/namea.file": true,
				"/nameb.file": true,
				"/namec.file": true,
				"/named.file": false,
			},
		},
		{
			pattern: "name[!abc].file",
			test: map[string]bool{
				"/named.file": true,
				"/namea.file": false,
				"/nameb.file": false,
				"/namec.file": false,
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.pattern, func(t *testing.T) {
			for v, expected := range tc.test {
				t.Run(v, func(t *testing.T) {
					b := Match(tc.pattern, v)
					require.Equal(t, expected, b, "expected %v for pattern %v and value %v", expected, tc.pattern, v)
				})
			}
		})
	}
}
