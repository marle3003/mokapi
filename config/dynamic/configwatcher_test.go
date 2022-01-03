package dynamic

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/test"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testcase struct {
	name       string
	filePath   string
	fn         func(t *testing.T, f *common.File)
	updatePath string
	updateFn   func(t *testing.T, f *common.File)
}

var testcases = []testcase{
	{
		name:     "openapi",
		filePath: "./test/openapi.yml",
		fn: func(t *testing.T, f *common.File) {
			assert.NotNil(t, f.Data)
			c := f.Data.(*openapi.Config)
			assert.Len(t, c.EndPoints, 1)
		},
		updatePath: "./test/openapi_update.yml",
		updateFn: func(t *testing.T, f *common.File) {
			assert.NotNil(t, f.Data)
			c := f.Data.(*openapi.Config)
			assert.Len(t, c.EndPoints, 2)
		},
	},
}

func TestWatcher(t *testing.T) {
	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Cleanup(func() { os.RemoveAll(tempDir) })
			w := NewConfigWatcher(&static.Config{
				Providers: static.Providers{
					File: static.FileProvider{Directory: tempDir}}})
			pool := safe.NewPool(context.Background())
			defer pool.Stop()
			ch := make(chan *common.File)
			w.AddListener(func(file *common.File) {
				ch <- file
			})
			err := w.Start(pool)
			test.Ok(t, err)

			_, err = createTempFile(tc.filePath, tempDir)
			test.Ok(t, err)

			timeout := time.After(2 * time.Second)
			select {
			case f := <-ch:
				tc.fn(t, f)
			case <-timeout:
				t.Fatal("timeout while waiting for file event")
			}

			if len(tc.updatePath) > 0 {
				_, err := createTempFile(tc.updatePath, tempDir)
				test.Ok(t, err)

				timeout := time.After(2 * time.Second)
				select {
				case f := <-ch:
					tc.updateFn(t, f)
				case <-timeout:
					t.Fatal("timeout while waiting for file event")
				}
			}
		})
	}
}

func TestWatcher_UpdateRef(t *testing.T) {
	tempDir := t.TempDir()
	pool := safe.NewPool(context.Background())
	t.Cleanup(func() {
		pool.Stop()
		os.RemoveAll(tempDir)
	})
	w := NewConfigWatcher(&static.Config{
		Providers: static.Providers{
			File: static.FileProvider{Directory: tempDir}}})

	ch := make(chan *common.File)
	err := w.Start(pool)
	test.Ok(t, err)

	file, err := createTempFile(testcases[0].filePath, tempDir)
	test.Ok(t, err)

	time.Sleep(time.Second)
	f, err := w.Read(mustParse(file), common.WithListener(func(file *common.File) {
		// send non-blocking to chan
		select {
		case ch <- file:
		default:
		}
	}))
	test.Ok(t, err)
	config, ok := f.Data.(*openapi.Config)
	test.IsTrue(t, ok)
	test.Equals(t, "test", config.Info.Name)
	assert.Len(t, config.EndPoints, 1)

	f1, err := createTempFile(testcases[0].updatePath, tempDir)
	test.Ok(t, err)
	test.Equals(t, file, f1)

	// wait for all events.
	timeout := time.After(3 * time.Second)
	gotEvent := false
Loop:
	for {
		select {
		case <-ch:
			gotEvent = true
		case <-timeout:
			break Loop
		}
	}
	assert.True(t, gotEvent)
	assert.Len(t, config.EndPoints, 2)
}

func mustParse(s string) *url.URL {
	s, err := filepath.Abs(s)
	if err != nil {
		panic(err)
	}
	u, err := url.Parse("file:" + s)
	if err != nil {
		panic(err)
	}
	return u
}

func createTempFile(srcPath string, destPath string) (string, error) {
	dest := filepath.Join(destPath, "tmp"+filepath.Ext(srcPath))
	file, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer file.Close()

	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()
	_, err = io.Copy(file, src)
	return dest, err
}
