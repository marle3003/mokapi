package npm

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file/filetest"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNpmProvider(t *testing.T) {
	root := "/"
	if filepath.Separator == '\\' {
		root = "C:\\"
	}

	mustTime := func(s string) time.Time {
		d, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}

	testcases := []struct {
		name string
		fs   *filetest.MockFS
		cfg  static.NpmProvider
		test func(t *testing.T, configs map[string]*dynamic.Config)
	}{
		{
			name: "node_modules in current directory and one file",
			fs: &filetest.MockFS{
				WorkingDir: root,
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:    "/node_modules/foo/foo.txt",
						IsDir:   false,
						Data:    []byte("foobar"),
						ModTime: mustTime("2024-01-02T15:04:05Z"),
					}}},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{{Name: "foo"}}},
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Contains(t, configs, "/node_modules/foo/foo.txt")
				require.Equal(t, []byte("foobar"), configs["/node_modules/foo/foo.txt"].Raw)
				require.Equal(t, mustTime("2024-01-02T15:04:05Z"), configs["/node_modules/foo/foo.txt"].Info.Time)
			},
		},
		{
			name: "node_modules in parent directory and one file",
			fs: &filetest.MockFS{
				WorkingDir: root + "foo",
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					}}},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{{Name: "foo"}}},
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Contains(t, configs, "/node_modules/foo/foo.txt")
				require.Equal(t, []byte("foobar"), configs["/node_modules/foo/foo.txt"].Raw)
			},
		},
		{
			name: "with global folder",
			fs: &filetest.MockFS{
				WorkingDir: root + "foo",
				Entries: []*filetest.Entry{
					{
						Name:  "/bar/node_modules",
						IsDir: true,
					},
					{
						Name:  "/bar/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/bar/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					}}},
			cfg: static.NpmProvider{
				GlobalFolders: []string{root + "bar/node_modules"},
				Packages:      []static.NpmPackage{{Name: "foo"}},
			},
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Contains(t, configs, "/bar/node_modules/foo/foo.txt")
				require.Equal(t, []byte("foobar"), configs["/bar/node_modules/foo/foo.txt"].Raw)
			},
		},
		{
			name: "node_modules in parent directory and two packages",
			fs: &filetest.MockFS{
				WorkingDir: root + "foo",
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					{
						Name:  "/node_modules/bar",
						IsDir: true,
					},
					{
						Name:  "/node_modules/bar/bar.txt",
						IsDir: false,
						Data:  []byte("bar"),
					},
				},
			},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{
				{Name: "foo"},
				{Name: "bar"},
			}},
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 2)
				require.Contains(t, configs, "/node_modules/foo/foo.txt")
				require.Equal(t, []byte("foobar"), configs["/node_modules/foo/foo.txt"].Raw)
				require.Contains(t, configs, "/node_modules/bar/bar.txt")
				require.Equal(t, []byte("bar"), configs["/node_modules/bar/bar.txt"].Raw)
			},
		},
		{
			name: "one module in current and one in parent",
			fs: &filetest.MockFS{
				WorkingDir: root + "foo",
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					{
						Name:  "/foo/node_modules/bar",
						IsDir: true,
					},
					{
						Name:  "/foo/node_modules/bar/bar.txt",
						IsDir: false,
						Data:  []byte("bar"),
					},
				},
			},
			cfg: static.NpmProvider{Packages: []static.NpmPackage{
				{Name: "foo"},
				{Name: "bar"},
			}},
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 2)
				require.Contains(t, configs, "/node_modules/foo/foo.txt")
				require.Equal(t, []byte("foobar"), configs["/node_modules/foo/foo.txt"].Raw)
				require.Contains(t, configs, "/foo/node_modules/bar/bar.txt")
				require.Equal(t, []byte("bar"), configs["/foo/node_modules/bar/bar.txt"].Raw)
			},
		},
		{
			name: "with allow list",
			fs: &filetest.MockFS{
				WorkingDir: root + "foo",
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					{
						Name:  "/node_modules/foo/bar.txt",
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
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Contains(t, configs, "/node_modules/foo/foo.txt")
				require.Equal(t, []byte("foobar"), configs["/node_modules/foo/foo.txt"].Raw)
			},
		},
		{
			name: "with include",
			fs: &filetest.MockFS{
				WorkingDir: root + "foo",
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					},
					{
						Name:  "/node_modules/foo/dist",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/dist/openapi",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/dist/openapi/foo.json",
						IsDir: false,
						Data:  []byte("{}"),
					},
					{
						Name:  "/node_modules/foo/dist/openapi/foo.txt",
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
			test: func(t *testing.T, configs map[string]*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Contains(t, configs, "/node_modules/foo/dist/openapi/foo.json")
				require.Equal(t, []byte("{}"), configs["/node_modules/foo/dist/openapi/foo.json"].Raw)
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

			ch := make(chan *dynamic.Config)
			p := NewFS(tc.cfg, tc.fs)
			err := p.Start(ch, pool)
			require.NoError(t, err)

			configs := map[string]*dynamic.Config{}
		Collect:
			for {
				select {
				case c := <-ch:
					path := c.Info.Inner().Url.Path
					if len(path) == 0 {
						path = c.Info.Inner().Url.Opaque
						path = strings.ReplaceAll(path, "\\", "/")
						path = strings.ReplaceAll(path, "C:/", "/")
					}
					configs[path] = c
				case <-time.After(time.Second):
					break Collect
				}
			}

			tc.test(t, configs)
		})
	}
}

func TestProvider_Read(t *testing.T) {
	root := "/"
	if filepath.Separator == '\\' {
		root = "C:\\"
	}

	testcases := []struct {
		name string
		fs   *filetest.MockFS
		cfg  static.NpmProvider
		test func(t *testing.T, p *Provider)
	}{
		{
			name: "simple npm package name",
			fs: &filetest.MockFS{
				WorkingDir: root,
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/foo/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					}}},
			test: func(t *testing.T, p *Provider) {
				u := mustUrl("npm://foo/foo.txt")
				c, err := p.Read(u)
				require.NoError(t, err)
				require.Equal(t, "foobar", string(c.Raw))
			},
		},
		{
			name: "scoped npm package name",
			fs: &filetest.MockFS{
				WorkingDir: root,
				Entries: []*filetest.Entry{
					{
						Name:  "/node_modules",
						IsDir: true,
					},
					{
						Name:  "/node_modules/@foo",
						IsDir: true,
					},
					{
						Name:  "/node_modules/@foo/bar",
						IsDir: true,
					},
					{
						Name:  "/node_modules/@foo/bar/foo.txt",
						IsDir: false,
						Data:  []byte("foobar"),
					}}},
			test: func(t *testing.T, p *Provider) {
				u := mustUrl("npm://bar/foo.txt?scope=@foo")
				c, err := p.Read(u)
				require.NoError(t, err)
				require.Equal(t, "foobar", string(c.Raw))
				require.Equal(t, "npm://bar/foo.txt?scope=@foo", c.Info.Url.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewFS(static.NpmProvider{}, tc.fs)
			tc.test(t, p)
		})
	}
}

func mustUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
