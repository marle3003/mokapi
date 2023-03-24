package file

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProvider_File(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"one file",
			func(t *testing.T) {
				ch := make(chan *common.Config)
				p := createFileProvider(t, "./test/openapi.yml")
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				err := p.Start(ch, pool)
				require.NoError(t, err)

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
				ch := make(chan *common.Config)
				files := []string{"./test/openapi.yml", "./test/openapi2.yml"}
				p := createFileProvider(t, files...)
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
		{
			"include */foo/*",
			func(t *testing.T) {
				ch := make(chan *common.Config)
				tempDir := t.TempDir()
				t.Cleanup(func() { os.RemoveAll(tempDir) })
				p := New(static.FileProvider{Directory: tempDir})
				err := createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "foo"))
				require.NoError(t, err)
				err = createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "foo/bar"))
				require.NoError(t, err)
				err = createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "bar"))
				require.NoError(t, err)
				p.Include = []string{"*/foo/*"}
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				err = p.Start(ch, pool)
				require.NoError(t, err)

				var got []string
			Stop:
				for {
					select {
					case c := <-ch:
						got = append(got, c.Url.String())
					case <-time.After(1 * time.Second):
						break Stop
					}
				}

				require.Len(t, got, 2)
			},
		},
		{
			"include */foo/*",
			func(t *testing.T) {
				ch := make(chan *common.Config)
				tempDir := t.TempDir()
				t.Cleanup(func() { os.RemoveAll(tempDir) })
				p := New(static.FileProvider{Directory: tempDir})
				err := createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "foo"))
				require.NoError(t, err)
				err = createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "foo/bar"))
				require.NoError(t, err)
				err = createTempFile("./test/openapi.yml", filepath.Join(p.cfg.Directory, "bar"))
				require.NoError(t, err)
				p.Include = []string{"*/foo/*"}
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				err = p.Start(ch, pool)
				require.NoError(t, err)

				var got []string
			Stop:
				for {
					select {
					case c := <-ch:
						got = append(got, c.Url.String())
					case <-time.After(1 * time.Second):
						break Stop
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

func createFileProvider(t *testing.T, files ...string) *Provider {
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	for _, file := range files {
		if len(file) > 0 {
			err := createTempFile(file, tempDir)
			require.NoError(t, err)
		}
	}

	p := New(static.FileProvider{Filename: strings.Join(files, string(os.PathListSeparator))})
	return p
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
