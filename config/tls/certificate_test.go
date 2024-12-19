package tls_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/tls"
	"os"
	"path"
	"testing"
)

func TestFileOrContent_Read(t *testing.T) {
	testcases := []struct {
		name        string
		content     tls.FileOrContent
		fileContent string
		test        func(t *testing.T, content tls.FileOrContent, fileContent string)
	}{
		{
			name:        "empty file",
			content:     "./foo.pem",
			fileContent: "",
			test: func(t *testing.T, content tls.FileOrContent, fileContent string) {
				temp := t.TempDir()
				err := os.WriteFile(path.Join(temp, "foo.pem"), []byte(fileContent), 0644)
				require.NoError(t, err)

				b, err := content.Read(temp)
				require.NoError(t, err)
				require.Equal(t, "", string(b))
			},
		},
		{
			name:        "file with content",
			content:     "./foo.pem",
			fileContent: "foo",
			test: func(t *testing.T, content tls.FileOrContent, fileContent string) {
				temp := t.TempDir()
				err := os.WriteFile(path.Join(temp, "foo.pem"), []byte(fileContent), 0644)
				require.NoError(t, err)

				b, err := content.Read(temp)
				require.NoError(t, err)
				require.Equal(t, "foo", string(b))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, tc.content, tc.fileContent)
		})
	}
}
