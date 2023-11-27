package npm

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file/filetest"
	"mokapi/config/static"
	"mokapi/safe"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNpmProvider(t *testing.T) {
	type file struct {
		path string
		data string
	}

	testcases := []struct {
		name string
		fs   *filetest.MockFS
		cfg  static.NpmProvider
		test func(t *testing.T, files map[string]file)
	}{
		{
			name: "node_modules in current directory and one file",
			fs: &filetest.MockFS{
				WorkingDir: "/",
				Entries: map[string]*filetest.Entry{
					"/node_modules": {
						Name:  "node_modules",
						IsDir: true,
					},
					"/node_modules/foo": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/foo/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					}}},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{{Name: "foo"}}},
			test: func(t *testing.T, files map[string]file) {
				require.Len(t, files, 1)
				require.Equal(t, "foobar", files["/node_modules/foo/foo.txt"].data)
			},
		},
		{
			name: "node_modules in parent directory and one file",
			fs: &filetest.MockFS{
				WorkingDir: "/foo",
				Entries: map[string]*filetest.Entry{
					"/node_modules": {
						Name:  "node_modules",
						IsDir: true,
					},
					"/node_modules/foo": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/foo/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					}}},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{{Name: "foo"}}},
			test: func(t *testing.T, files map[string]file) {
				require.Len(t, files, 1)
				require.Equal(t, "foobar", files["/node_modules/foo/foo.txt"].data)
			},
		},
		{
			name: "node_modules in parent directory and two packages",
			fs: &filetest.MockFS{
				WorkingDir: "/foo",
				Entries: map[string]*filetest.Entry{
					"/node_modules": {
						Name:  "node_modules",
						IsDir: true,
					},
					"/node_modules/foo": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/foo/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					"/node_modules/bar": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/bar/bar.txt": {
						Name:  "bar.txt",
						IsDir: false,
						Data:  []byte("bar"),
					},
				},
			},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{
				{Name: "foo"},
				{Name: "bar"},
			}},
			test: func(t *testing.T, files map[string]file) {
				require.Len(t, files, 2)
				require.Equal(t, "foobar", files["/node_modules/foo/foo.txt"].data)
				require.Equal(t, "bar", files["/node_modules/bar/bar.txt"].data)
			},
		},
		{
			name: "one module in current and one in parent",
			fs: &filetest.MockFS{
				WorkingDir: "/foo",
				Entries: map[string]*filetest.Entry{
					"/node_modules": {
						Name:  "node_modules",
						IsDir: true,
					},
					"/node_modules/foo": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/foo/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					"/foo/node_modules/bar": {
						Name:  "foo",
						IsDir: true,
					},
					"/foo/node_modules/bar/bar.txt": {
						Name:  "bar.txt",
						IsDir: false,
						Data:  []byte("bar"),
					},
				},
			},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{
				{Name: "foo"},
				{Name: "bar"},
			}},
			test: func(t *testing.T, files map[string]file) {
				require.Len(t, files, 2)
				require.Equal(t, "foobar", files["/node_modules/foo/foo.txt"].data)
				require.Equal(t, "bar", files["/foo/node_modules/bar/bar.txt"].data)
			},
		},
		{
			name: "with allow list",
			fs: &filetest.MockFS{
				WorkingDir: "/foo",
				Entries: map[string]*filetest.Entry{
					"/node_modules": {
						Name:  "node_modules",
						IsDir: true,
					},
					"/node_modules/foo": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/foo/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					"/node_modules/foo/bar.txt": {
						Name:  "bar.txt",
						IsDir: false,
						Data:  []byte("bar"),
					},
				},
			},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{
				{
					Name:  "foo",
					Files: []string{"foo.txt"},
				},
			}},
			test: func(t *testing.T, files map[string]file) {
				require.Len(t, files, 1)
				require.Equal(t, "foobar", files["/node_modules/foo/foo.txt"].data)
			},
		},
		{
			name: "with include",
			fs: &filetest.MockFS{
				WorkingDir: "/foo",
				Entries: map[string]*filetest.Entry{
					"/node_modules": {
						Name:  "node_modules",
						IsDir: true,
					},
					"/node_modules/foo": {
						Name:  "foo",
						IsDir: true,
					},
					"/node_modules/foo/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					"/node_modules/foo/dist": {
						Name:  "dist",
						IsDir: true,
					},
					"/node_modules/foo/dist/openapi": {
						Name:  "openapi",
						IsDir: true,
					},
					"/node_modules/foo/dist/openapi/foo.json": {
						Name:  "foo.json",
						IsDir: false,
						Data:  []byte("{}"),
					},
					"/node_modules/foo/dist/openapi/foo.txt": {
						Name:  "foo.txt",
						IsDir: false,
						Data:  []byte("bar"),
					},
				},
			},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{
				{
					Name:    "foo",
					Include: []string{"dist/**/*.json"},
				},
			}},
			test: func(t *testing.T, files map[string]file) {
				require.Len(t, files, 1)
				require.Equal(t, "{}", files["/node_modules/foo/dist/openapi/foo.json"].data)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pool := safe.NewPool(context.Background())
			t.Cleanup(func() {
				pool.Stop()
			})

			ch := make(chan *common.Config)
			p := NewFS(tc.cfg, tc.fs)
			err := p.Start(ch, pool)
			require.NoError(t, err)

			files := map[string]file{}
		Collect:
			for {
				select {
				case c := <-ch:
					path := c.Info.Url.Path
					if len(path) == 0 {
						path = c.Info.Url.Opaque
						parent, _ := os.Getwd()
						path = strings.Replace(path, parent+"\\", "", 1)
					}
					files[path] = file{path, string(c.Raw)}
				case <-time.After(time.Second):
					break Collect
				}
			}

			tc.test(t, files)
		})
	}
}
