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
Walk:
	for path, f := range m.fs {
		for _, dir := range ignoreDirs {
			if strings.HasPrefix(path, dir) {
				continue Walk
			}
		}
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
	testcases := []struct {
		name string
		fs   *mockFS
		cfg  static.FileProvider
		test func(t *testing.T, files []string)
	}{
		{
			name: "one file",
			fs: &mockFS{map[string]*entry{"foo.txt": {
				name:  "foo.txt",
				isDir: false,
				data:  []byte("foobar"),
			}}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []string) {
				require.Len(t, files, 1)
				require.Equal(t, "foo.txt", filepath.Base(files[0]))
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
			test: func(t *testing.T, files []string) {
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
			test: func(t *testing.T, files []string) {
				require.Len(t, files, 1)
				require.Equal(t, "_foo.txt", filepath.Base(files[0]))
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
			test: func(t *testing.T, files []string) {
				require.Len(t, files, 1)
				require.Equal(t, "foo.txt", filepath.Base(files[0]))
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
			test: func(t *testing.T, files []string) {
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
			test: func(t *testing.T, files []string) {
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
			test: func(t *testing.T, files []string) {
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
			test: func(t *testing.T, files []string) {
				require.Len(t, files, 0)
			},
		},
		{
			name: "mokapiignore all files with re-include",
			fs: &mockFS{map[string]*entry{
				".mokapiignore": {
					name:  ".mokapiignore",
					isDir: false,
					data:  []byte("/*\n!/dir"),
				},
				"/bar.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
				"dir/foo.txt": {
					name:  "foo.txt",
					isDir: false,
					data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directory: "./"},
			test: func(t *testing.T, files []string) {
				require.Len(t, files, 1)
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
			var files []string
		Collect:
			for {
				select {
				case c := <-ch:
					files = append(files, c.Url.Path)
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
