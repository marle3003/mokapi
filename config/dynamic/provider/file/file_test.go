package file

import (
	"context"
	"io"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/test"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProvider(t *testing.T) {
	ch := make(chan *common.Config)
	p := createProvider(t, "./test/openapi.yml")
	pool := safe.NewPool(context.Background())
	defer pool.Stop()
	err := p.Start(ch, pool)
	test.Ok(t, err)

	timeout := time.After(time.Second)
	select {
	case c := <-ch:
		test.Assert(t, len(c.Url.String()) > 0, "url is set")
		test.Assert(t, len(c.Raw) > 0, "got data")
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}
}

func TestWatch(t *testing.T) {
	ch := make(chan *common.Config)
	p := createProvider(t, "")
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	err := p.Start(ch, pool)
	test.Ok(t, err)

	time.Sleep(500 * time.Millisecond)
	err = createTempFile("./test/openapi.yml", p.cfg.Directory)
	test.Ok(t, err)

	timeout := time.After(2 * time.Second)
	select {
	case c := <-ch:
		test.Assert(t, len(c.Url.String()) > 0, "url is set")
		test.Assert(t, len(c.Raw) > 0, "got data")
	case <-timeout:
		t.Fatal("timeout while waiting for file event")
	}
}

func createProvider(t *testing.T, file string) *Provider {
	tempDir := t.TempDir()
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	if len(file) > 0 {
		err := createTempFile(file, tempDir)
		test.Ok(t, err)
	}

	p := New(static.FileProvider{Directory: tempDir})
	return p
}

func createTempFile(srcPath string, destPath string) error {
	file, err := os.CreateTemp(destPath, filepath.Ext(srcPath))
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
