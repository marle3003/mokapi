package file

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file/filetest"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestProvider(t *testing.T) {
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
		cfg  static.FileProvider
		test func(t *testing.T, configs []*dynamic.Config)
	}{
		{
			name: "not in same dir",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "bar/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				}}},
			cfg: static.FileProvider{Directories: []string{"./foo"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "one file",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:    "foo.txt",
					IsDir:   false,
					Data:    []byte("foobar"),
					ModTime: mustTime("2024-01-02T15:04:05Z"),
				}}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Equal(t, "foo.txt", filepath.Base(configs[0].Info.Path()))
				require.Equal(t, []byte("foobar"), configs[0].Raw)
				require.Equal(t, mustTime("2024-01-02T15:04:05Z"), configs[0].Info.Time)
			},
		},
		{
			name: "file no UTF-8-BOM",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "foo.txt",
					IsDir: false,
					Data:  []byte("fo"),
				}}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Equal(t, "foo.txt", filepath.Base(configs[0].Info.Path()))
				require.Equal(t, []byte("fo"), configs[0].Raw)
			},
		},
		{
			name: "file UTF-8-BOM",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "foo.txt",
					IsDir: false,
					Data:  []byte{0xEF, 0xBB, 0xBF, 'f', 'o', 'o', 'b', 'a', 'r'},
				}}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Equal(t, "foo.txt", filepath.Base(configs[0].Info.Path()))
				require.Equal(t, []byte("foobar"), configs[0].Raw)
			},
		},
		{
			name: "skipped file",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "_foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				}}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "custom skip file overwrites default skip",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "$foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "_foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}, SkipPrefix: []string{"$"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Equal(t, "_foo.txt", filepath.Base(configs[0].Info.Path()))
			},
		},
		{
			name: "file in directory",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "dir",
					IsDir: true,
				},
				{
					Name:  "dir/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Equal(t, "foo.txt", filepath.Base(configs[0].Info.Path()))
			},
		},
		{
			name: "file in skipped directory",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "_dir",
					IsDir: true,
				},
				{
					Name:  "_dir/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "mokapiignore",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("foo.txt"),
				},
				{
					Name:  "foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "mokapiignore in subfolder",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "dir",
					IsDir: true,
				},
				{
					Name:  "dir/.mokapiignore",
					IsDir: false,
					Data:  []byte("foo.txt"),
				},
				{
					Name:  "dir/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "mokapiignore in subfolder overrides parent",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("foo.txt"),
				},
				{
					Name:  "dir",
					IsDir: true,
				},
				{
					Name:  "dir/.mokapiignore",
					IsDir: false,
					Data:  []byte(""),
				},
				{
					Name:  "dir/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.Equal(t, "foobar", string(configs[0].Raw))
			},
		},
		{
			name: "mokapiignore but re-include",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("*.txt\n!dir/foo.txt"),
				},
				{
					Name:  "dir/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
			},
		},
		{
			name: "mokapiignore with re-include but excluding again",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("dir\n!dir/foo.txt\ndir"),
				},
				{
					Name:  "dir/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "ignoring all files but re-include some",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("/*\n!/dir/bar"),
				},
				{
					Name:  "/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "/dir/bar/foo.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
			},
		},
		{
			name: "mokapiignore ignore all but js files",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("**/*.*\n!*.js"),
				},
				{
					Name:  "/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "dir/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "dir/foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
				require.True(t, strings.HasSuffix(configs[0].Info.Path(), filepath.Join("dir", "foo.js")))
			},
		},
		{
			name: "mokapiignore ignore all but js and ts files",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("**/*.*\n!*.js\n!mokapi.ts"),
				},
				{
					Name:  "/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "dir/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "dir/foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "dir/mokapi.ts",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 2)
				require.True(t, strings.HasSuffix(configs[0].Info.Path(), filepath.Join("dir", "foo.js")))
				require.True(t, strings.HasSuffix(configs[1].Info.Path(), filepath.Join("dir", "mokapi.ts")))
			},
		},
		{
			name: "mokapiignore all files but specific sub folder",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte("**/*.*\n!/foo/bar/**"),
				},
				{
					Name:  "/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "dir/bar.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "foo/bar/foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "foo/bar/dir/foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 2)
				sort.Slice(configs, func(i, j int) bool {
					return configs[i].Info.Path() < configs[j].Info.Path()
				})
				require.True(t, strings.HasSuffix(configs[0].Info.Path(), filepath.Join("bar", "dir", "foo.js")), "%v does not match suffix %v", configs[1].Info.Path(), filepath.Join("bar", "dir", "foo.js"))
				require.True(t, strings.HasSuffix(configs[1].Info.Path(), filepath.Join("bar", "foo.js")), "%v does not match suffix %v", configs[0].Info.Path(), filepath.Join("bar", "foo.js"))
			},
		},
		{
			name: ".git",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  ".mokapiignore",
					IsDir: false,
					Data:  []byte(".git/"),
				},
				{
					Name:  ".git/foo",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 0)
			},
		},
		{
			name: "include",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "foo",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "bar",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}, Include: []string{"foo"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 1)
			},
		},
		{
			name: "include with folders",
			fs: &filetest.MockFS{Entries: []*filetest.Entry{
				{
					Name:  "foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "foo",
					IsDir: true,
					Data:  []byte("foobar"),
				},
				{
					Name:  "foo/foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "foo/bar.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "bar",
					IsDir: true,
					Data:  []byte("foobar"),
				},
				{
					Name:  "bar/foo.js",
					IsDir: false,
					Data:  []byte("foobar"),
				},
				{
					Name:  "bar/test.txt",
					IsDir: false,
					Data:  []byte("foobar"),
				},
			}},
			cfg: static.FileProvider{Directories: []string{"./"}, Include: []string{"*.js"}},
			test: func(t *testing.T, configs []*dynamic.Config) {
				require.Len(t, configs, 4)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := NewWithWalker(tc.cfg, tc.fs)
			pool := safe.NewPool(context.Background())
			t.Cleanup(func() {
				pool.Stop()
			})
			ch := make(chan *dynamic.Config)
			err := p.Start(ch, pool)
			require.NoError(t, err)
			var configs []*dynamic.Config
		Collect:
			for {
				select {
				case c := <-ch:
					configs = append(configs, c)
				case <-time.After(2 * time.Second):
					break Collect
				}
			}
			tc.test(t, configs)
			close(ch)
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
					require.True(t, len(c.Info.Url.String()) > 0, "url is set")
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
						got = append(got, c.Info.Url.String())
						require.True(t, len(c.Info.Url.String()) > 0, "url is set")
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
				ch := make(chan *dynamic.Config)
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
						got = append(got, c.Info.Url.String())
						require.True(t, len(c.Info.Url.String()) > 0, "url is set")
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
	ch := make(chan *dynamic.Config)
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	p := New(static.FileProvider{Directories: []string{tempDir}})
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	err := p.Start(ch, pool)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	err = createTempFile("./test/openapi.yml", p.cfg.Directories[0])
	require.NoError(t, err)

	timeout := time.After(5 * time.Second)
	select {
	case c := <-ch:
		require.True(t, len(c.Info.Url.String()) > 0, "url is set")
		require.True(t, len(c.Raw) > 0, "got data")
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}
}

func TestWatch_Create_SubFolder_And_Add_File(t *testing.T) {
	ch := make(chan *dynamic.Config)
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	p := New(static.FileProvider{Directories: []string{tempDir}})
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	err := p.Start(ch, pool)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	err = createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directories[0], "foo"))
	require.NoError(t, err)

	timeout := time.After(5 * time.Second)
	select {
	case c := <-ch:
		require.True(t, len(c.Info.Url.String()) > 0, "url is set")
		require.True(t, len(c.Raw) > 0, "got data")
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}
}

func TestWatch_UpdateFile_When_Skipped_But_Referenced(t *testing.T) {
	ch := make(chan *dynamic.Config)
	defer close(ch)

	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	p := New(static.FileProvider{Directories: []string{tempDir}})
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	err := p.Start(ch, pool)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)

	err = createFile(filepath.Join(tempDir, "foo.js"), "foo")
	require.NoError(t, err)
	waitFileUpdate(t, ch)
	err = createFile(filepath.Join(tempDir, "_bar.skip"), "bar")
	u, err := url.Parse("file:" + filepath.Join(tempDir, "_bar.skip"))
	require.NoError(t, err)
	_, err = p.Read(u)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "_bar.skip"), []byte("update"), 0700)
	require.NoError(t, err)

	file := waitFileUpdate(t, ch)
	require.NotNil(t, file)
	require.Equal(t, "_bar.skip", filepath.Base(file.Info.Path()))
}

func createAndStartFileProvider(t *testing.T, files ...string) (*Provider, chan *dynamic.Config) {
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	for _, file := range files {
		if len(file) > 0 {
			err := createTempFile(file, tempDir)
			require.NoError(t, err)
		}
	}

	p := New(static.FileProvider{Filenames: []string{strings.Join(files, string(os.PathListSeparator))}})
	pool := safe.NewPool(context.Background())
	t.Cleanup(func() {
		pool.Stop()
	})
	ch := make(chan *dynamic.Config)
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

	p := New(static.FileProvider{Directories: []string{strings.Join(dirs, string(os.PathListSeparator))}})
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

func createFile(path, content string) error {
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Write([]byte(content))
	file.Close()

	return err
}

func waitFileUpdate(t *testing.T, ch chan *dynamic.Config) *dynamic.Config {
	timeout := time.After(5 * time.Second)
	select {
	case c := <-ch:
		return c
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}

	return nil
}
