package file

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"io/fs"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type entry struct {
	name  string
	isDir bool
	data  []byte
}

type mockFS struct {
	fs map[string]*entry
}

func (m *mockFS) ReadFile(path string) ([]byte, error) {
	if f, ok := m.fs[path]; ok {
		return f.data, nil
	}
	return nil, fmt.Errorf("not foubnd")
}

func (m *mockFS) Walk(root string, visit fs.WalkDirFunc) error {
	var ignoreDirs []string

	// loop order in a map is not safe, order path to ensure SkipDir
	keys := make([]string, 0, len(m.fs))
	for k := range m.fs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
Walk:
	for _, path := range keys {
		for _, dir := range ignoreDirs {
			if strings.HasPrefix(path, dir) {
				continue Walk
			}
		}
		f := m.fs[path]
		if err := visit(path, f, nil); err == fs.SkipDir {
			ignoreDirs = append(ignoreDirs, path)
		}
	}
	return nil
}

func (e *entry) Name() string {
	return e.name
}

func (e *entry) IsDir() bool {
	return e.isDir
}

func (e *entry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}

func (e *entry) Info() (fs.FileInfo, error) {
	return nil, nil
}

func TestProvider(t *testing.T) {
	type file struct {
		path string
		data string
	}
	testcases := []struct {
		name string
		fs   *mockFS
		cfg  static.FileProvider
		test func(t *testing.T, files []file)
	}{
		{
			name: "one file",
			fs: &mockFS{map[string]*entry{"foo.txt": {
				name:  "foo.txt",
				isDir: false,
				data:  []byte("foobar"),
			}}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
				require.Equal(t, "foo.txt", filepath.Base(files[0].path))
				require.Equal(t, "foobar", files[0].data)
			},
		},
		{
			name: "file UTF-8-BOM",
			fs: &mockFS{map[string]*entry{"foo.txt": {
				name:  "foo.txt",
				isDir: false,
				data:  []byte{0xEF, 0xBB, 0xBF, 'f', 'o', 'o', 'b', 'a', 'r'},
			}}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
				require.Equal(t, "foo.txt", filepath.Base(files[0].path))
				require.Equal(t, "foobar", files[0].data)
			},
		},
		{
			name: "skipped file",
			fs: &mockFS{map[string]*entry{"_foo.txt": {
				name:  "_foo.txt",
				isDir: false,
				data:  []byte("foobar"),
			}}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 0)
			},
		},
		{
			name: "custom skip file overwrites default skip",
			fs: &mockFS{map[string]*entry{
				"$foo.txt": {
					name:  "$foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"_foo.txt": {
					name:  "_foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./", SkipPrefix: []string{"$"}},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
				require.Equal(t, "_foo.txt", filepath.Base(files[0].path))
			},
		},
		{
			name: "file in directory",
			fs: &mockFS{map[string]*entry{
				"dir": {
					name:  "dir",
					isDir: true,
				},
				"dir/foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
				require.Equal(t, "foo.txt", filepath.Base(files[0].path))
			},
		},
		{
			name: "file in skipped directory",
			fs: &mockFS{map[string]*entry{
				"_dir": {
					name:  "_dir",
					isDir: true,
				},
				"_dir/foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 0)
			},
		},
		{
			name: "mokapiignore",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("foo.txt"),
				},
				"foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 0)
			},
		},
		{
			name: "mokapiignore but re-include",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("*.txt\n!dir/foo.txt"),
				},
				"dir/foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
			},
		},
		{
			name: "mokapiignore with re-include but parent is blocked",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("dir\n!dir/foo.txt"),
				},
				"dir/foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 0)
			},
		},
		{
			name: "mokapiignore all files with re-include",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("/*\n!/dir/bar"),
				},
				"/bar.txt": {
					name:  "bar.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"/dir/bar/foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
			},
		},
		{
			name: "mokapiignore only js files",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("**/*.*\n!*.js"),
				},
				"/bar.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"dir/bar.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"dir/foo.js": {
					name:  "foo.js",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 1)
				require.True(t, strings.HasSuffix(files[0].path, filepath.Join("dir", "foo.js")))
			},
		},
		{
			name: "mokapiignore all files but specific sub folder",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("**/*.*\n!/foo/bar/**"),
				},
				"/bar.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"dir/bar.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"foo/bar/foo.js": {
					name:  "foo.js",
					isDir: false,
					data:  []byte("foobar"),
				},
				"foo/bar/dir/foo.js": {
					name:  "foo.js",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []file) {
				require.Len(t, files, 2)
				sort.Slice(files, func(i, j int) bool {
					return files[i].path < files[j].path
				})
				require.True(t, strings.HasSuffix(files[0].path, filepath.Join("bar", "dir", "foo.js")), "%v does not match suffix %v", files[1].path, filepath.Join("bar", "dir", "foo.js"))
				require.True(t, strings.HasSuffix(files[1].path, filepath.Join("bar", "foo.js")), "%v does not match suffix %v", files[0].path, filepath.Join("bar", "foo.js"))
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := NewWithWalker(tc.cfg, tc.fs)
			pool := safe.NewPool(context.Background())
			t.Cleanup(func() {
				pool.Stop()
			})
			ch := make(chan *common.Config)
			err := p.Start(ch, pool)
			require.NoError(t, err)
			var files []file
		Collect:
			for {
				select {
				case c := <-ch:
					path := c.Url.Path
					if len(path) == 0 {
						path = c.Url.Opaque
						parent, _ := os.Getwd()
						path = strings.Replace(path, parent+"\\", "", 1)
					}
					files = append(files, file{path, string(c.Raw)})
				case <-time.After(time.Second):
					break Collect
				}
			}
			tc.test(t, files)
		})
	}
}

func TestProvider_File(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"one file",
			func(t *testing.T) {
				_, ch := createAndStartFileProvider(t, "./test/openapi.yml")
				timeout := time.After(time.Second)
				select {
				case c := <-ch:
					require.True(t, len(c.Url.String()) > 0, "url is set")
					require.True(t, len(c.Raw) > 0, "got data")
				case <-timeout:
					t.Fatal("timeout while waiting for file event")
				}
			},
		},
		{
			"two file",
			func(t *testing.T) {
				files := []string{"./test/openapi.yml", "./test/openapi2.yml"}
				_, ch := createAndStartFileProvider(t, files...)

				var got []string
				timeout := time.After(5 * time.Second)
				for i := 0; i < 2; i++ {
					select {
					case c := <-ch:
						got = append(got, c.Url.String())
						require.True(t, len(c.Url.String()) > 0, "url is set")
						require.True(t, len(c.Raw) > 0, "got data")
					case <-timeout:
						t.Fatal("timeout while waiting for file event")
					}
				}

				require.Len(t, got, 2)
			},
		},
		{
			"two dirs",
			func(t *testing.T) {
				ch := make(chan *common.Config)
				files := []string{"./test/openapi.yml", "./test/openapi2.yml"}
				p := createDirectoryProvider(t, files...)
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				err := p.Start(ch, pool)
				require.NoError(t, err)

				var got []string
				timeout := time.After(5 * time.Second)
				for i := 0; i < 2; i++ {
					select {
					case c := <-ch:
						got = append(got, c.Url.String())
						require.True(t, len(c.Url.String()) > 0, "url is set")
						require.True(t, len(c.Raw) > 0, "got data")
					case <-timeout:
						t.Fatal("timeout while waiting for file event")
					}
				}

				require.Len(t, got, 2)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}

}

func TestWatch_AddFile(t *testing.T) {
	ch := make(chan *common.Config)
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	p := New(static.FileProvider{Directory: tempDir})
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	err := p.Start(ch, pool)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	err = createTempFile("./test/openapi.yml", p.cfg.Directory)
	require.NoError(t, err)

	timeout := time.After(5 * time.Second)
	select {
	case c := <-ch:
		require.True(t, len(c.Url.String()) > 0, "url is set")
		require.True(t, len(c.Raw) > 0, "got data")
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}
}

func TestWatch_Create_SubFolder_And_Add_File(t *testing.T) {
	ch := make(chan *common.Config)
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	p := New(static.FileProvider{Directory: tempDir})
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	err := p.Start(ch, pool)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	err = createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "foo"))
	require.NoError(t, err)

	timeout := time.After(5 * time.Second)
	select {
	case c := <-ch:
		require.True(t, len(c.Url.String()) > 0, "url is set")
		require.True(t, len(c.Raw) > 0, "got data")
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}
}

func createAndStartFileProvider(t *testing.T, files ...string) (*Provider, chan *common.Config) {
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	for _, file := range files {
		if len(file) > 0 {
			err := createTempFile(file, tempDir)
			require.NoError(t, err)
		}
	}

	p := New(static.FileProvider{Filename: strings.Join(files, string(os.PathListSeparator))})
	pool := safe.NewPool(context.Background())
	t.Cleanup(func() {
		pool.Stop()
	})
	ch := make(chan *common.Config)
	err := p.Start(ch, pool)
	require.NoError(t, err)
	return p, ch
}

func createDirectoryProvider(t *testing.T, files ...string) *Provider {
	var dirs []string
	for _, file := range files {
		tempDir := t.TempDir()
		dirs = append(dirs, tempDir)
		t.Cleanup(func() { os.RemoveAll(tempDir) })

		if len(file) > 0 {
			err := createTempFile(file, tempDir)
			require.NoError(t, err)
		}
	}

	p := New(static.FileProvider{Directory: strings.Join(dirs, string(os.PathListSeparator))})
	return p
}

func createTempFile(srcPath string, destPath string) error {
	err := os.MkdirAll(destPath, 0700)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(destPath, "*"+filepath.Ext(srcPath))
	if err != nil {
		return err
	}
	defer file.Close()

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	_, err = io.Copy(file, src)

	return err
}
