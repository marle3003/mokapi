package main

import (
	"bytes"
	"io"
	"mokapi/version"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain_Skeleton(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		test func(t *testing.T, out string)
	}{
		{
			name: "version",
			args: []string{"--version"},
			test: func(t *testing.T, out string) {
				require.Equal(t, "1.0\n", out)
			},
		},
		{
			name: "version short",
			args: []string{"-v"},
			test: func(t *testing.T, out string) {
				require.Equal(t, "1.0\n", out)
			},
		},
		{
			name: "help",
			args: []string{"--help"},
			test: func(t *testing.T, out string) {
				require.Contains(t, out, "Mokapi is an easy, modern and flexible API mocking tool using Go")
			},
		},
		{
			name: "help short",
			args: []string{"-h"},
			test: func(t *testing.T, out string) {
				require.Contains(t, out, "Mokapi is an easy, modern and flexible API mocking tool using Go")
			},
		},
		{
			name: "generate-cli-skeleton",
			args: []string{"--generate-cli-skeleton"},
			test: func(t *testing.T, out string) {
				require.Equal(t,
					`log:
    level: info
    format: text
providers:
    file:
        filenames: []
        directories: []
        skipPrefix:
            - _
        include: []
    git:
        urls: []
        pullInterval: ""
        tempDir: ""
        repositories: []
    http:
        urls: []
        pollInterval: ""
        pollTimeout: ""
        proxy: ""
        tlsSkipVerify: false
        ca: ""
    npm:
        globalFolders: []
        packages: []
api:
    port: 8080
    path: ""
    base: ""
    dashboard: true
    search:
        enabled: true
rootCaCert: ""
rootCaKey: ""
configs: []
event:
    store:
        default:
            size: 100
certificates:
    static: []
data-gen:
    optionalProperties: "0.85"
`, out)
			},
		},
		{
			name: "generate-cli-skeleton providers",
			args: []string{"--generate-cli-skeleton", "providers"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `file:
    filenames: []
    directories: []
    skipPrefix:
        - _
    include: []
git:
    urls: []
    pullInterval: ""
    tempDir: ""
    repositories: []
http:
    urls: []
    pollInterval: ""
    pollTimeout: ""
    proxy: ""
    tlsSkipVerify: false
    ca: ""
npm:
    globalFolders: []
    packages: []
`, out)
			},
		},
	}

	stdOut := os.Stdout
	stdErr := os.Stderr

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			v := version.BuildVersion
			version.BuildVersion = "1.0"
			defer func() {
				version.BuildVersion = v
			}()

			reader, writer, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = writer
			os.Stderr = writer
			defer func() {
				os.Stdout = stdOut
				os.Stderr = stdErr
			}()

			os.Args = nil
			os.Args = append(os.Args, "mokapi.exe")
			os.Args = append(os.Args, tc.args...)

			main()

			_ = writer.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)
			_ = reader.Close()

			tc.test(t, buf.String())
		})
	}
}
