package static_test

import (
	"encoding/json"
	"mokapi/config/static"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYaml(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "directories as string array",
			data: `
providers:
  file:
    directories:
      - ./dir
`,
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.FileConfig{{Path: "./dir"}}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories as file config",
			data: `
providers:
  file:
    directories:
      - path: ./dir
`,
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.FileConfig{{Path: "./dir"}}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories as mixed",
			data: `
providers:
  file:
    directories:
      - path: ./dir
        include: ['*.js']
      - ./foo
`,
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.FileConfig{
					{Path: "./dir", Include: []string{"*.js"}},
					{Path: "./foo"},
				}, cfg.Providers.File.Directories)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *static.Config
			err := yaml.Unmarshal([]byte(tc.data), &cfg)
			require.NoError(t, err)
			tc.test(t, cfg)
		})
	}
}

func TestJson(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "directories as string array",
			data: `{
"providers": {
  "file": {
    "directories": ["./dir"]
  }
}}
`,
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.FileConfig{{Path: "./dir"}}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories as file config",
			data: `{
"providers": {
  "file": {
    "directories": [{"path":"./dir"}]
  }
}}
`,
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.FileConfig{{Path: "./dir"}}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories as mixed",
			data: `{
"providers": {
  "file": {
    "directories": [{"path":"./dir","include":["*.js"]}, "./foo"]
  }
}}
`,
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.FileConfig{
					{Path: "./dir", Include: []string{"*.js"}},
					{Path: "./foo"},
				}, cfg.Providers.File.Directories)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *static.Config
			err := json.Unmarshal([]byte(tc.data), &cfg)
			require.NoError(t, err)
			tc.test(t, cfg)
		})
	}
}
