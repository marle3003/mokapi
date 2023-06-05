package file

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMatch(t *testing.T) {
	testcases := []struct {
		name    string
		pattern string
		test    map[string]bool
	}{
		{
			name:    "name",
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
			name:    "name/",
			pattern: "name/",
			test: map[string]bool{
				"/name/file.txt":     true,
				"/name/log/name.log": true,
				"/name.log":          false,
			},
		},
		{
			name:    "name.file",
			pattern: "name.file",
			test: map[string]bool{
				"/name.file":     true,
				"/lib/name.file": true,
				"/name.log":      false,
			},
		},
		{
			name:    "/name.file",
			pattern: "/name.file",
			test: map[string]bool{
				"/name.file":     true,
				"name.file":      true,
				"/lib/name.file": false,
			},
		},
		{
			name:    "lib/name.file",
			pattern: "lib/name.file",
			test: map[string]bool{
				"/lib/name.file":      true,
				"name.file":           false,
				"/test/lib/name.file": false,
			},
		},
		{
			name:    "double star at start with folder",
			pattern: "**/lib/name.file",
			test: map[string]bool{
				"/lib/name.file":      true,
				"/test/lib/name.file": true,
				"name.file":           false,
			},
		},
		{
			name:    "double star at start",
			pattern: "**/name",
			test: map[string]bool{
				"/name/log.file":     true,
				"/lib/name/log.file": true,
				"/name/lib/log.file": true,
			},
		},
		{
			name:    "double star in middle",
			pattern: "/lib/**/name",
			test: map[string]bool{
				"/lib/test/ver1/name/log.file": true,
				"/lib/name/log.file":           true,
				"/name/lib/log.file":           false,
			},
		},
		{
			name:    "*.file",
			pattern: "*.file",
			test: map[string]bool{
				"/name.file":     true,
				"name.file":      true,
				"/lib/name.file": true,
				"name.txt":       false,
			},
		},
		{
			name:    "/*",
			pattern: "/*",
			test: map[string]bool{
				"/name.file":     true,
				"name.file":      true,
				"/lib/name.file": false,
				"lib/name.file":  false,
				"name.txt":       true,
			},
		},
		{
			name:    "*name/",
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
			name:    "*/name/",
			pattern: "*/name/",
			test: map[string]bool{
				"/foo/name/log.file": true,
				"/name/log.file":     false,
				"/foo/lib/name.file": false,
				"name.txt":           false,
			},
		},
		{
			name:    "name?.file",
			pattern: "name?.file",
			test: map[string]bool{
				"/names.file":  true,
				"/name1.file":  true,
				"/names1.file": false,
			},
		},
		{
			name:    "name[a-z].file",
			pattern: "name[a-z].file",
			test: map[string]bool{
				"/names.file":  true,
				"/nameb.file":  true,
				"/name1.file":  false,
				"/names1.file": false,
			},
		},
		{
			name:    "name[abc].file",
			pattern: "name[abc].file",
			test: map[string]bool{
				"/namea.file": true,
				"/nameb.file": true,
				"/namec.file": true,
				"/named.file": false,
			},
		},
		{
			name:    "name[!abc].file",
			pattern: "name[!abc].file",
			test: map[string]bool{
				"/named.file": true,
				"/namea.file": false,
				"/nameb.file": false,
				"/namec.file": false,
			},
		},
		{
			name:    "foo/*",
			pattern: "foo/*",
			test: map[string]bool{
				"/foo/test.json":   true,
				"/foo/bar":         true,
				"/foo/bar/hello.c": false,
			},
		},
		{
			name:    "double star at end",
			pattern: "/foo/bar/**",
			test: map[string]bool{
				"/foo/test.json":       false,
				"/foo/bar":             false,
				"/foo/bar/hello.c":     true,
				"/foo/bar/dir/hello.c": true,
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for v, expected := range tc.test {
				t.Run(v, func(t *testing.T) {
					b := Match(tc.pattern, v)
					require.Equal(t, expected, b, "expected %v for pattern %v and value %v", expected, tc.pattern, v)
				})
			}
		})
	}
}
