package dynamic

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestConfigWatcher_Openapi(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"update referenced file",
			func(t *testing.T) {
				root := `
openapi: 3.0.1
info:
  title: "update referenced file"
paths:
  /users:
    $ref: 'paths.yml#/paths/users'`
				// referenced files have to contain the header.
				// If it is missing, updates will not work
				path := `
paths:
  /users:
    get:
      summary: "foo"`

				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				p := &testproviderMap{
					files: map[string]*common.Config{
						"/root.yml": {
							Info:     common.ConfigInfo{Url: mustParse("/root.yml")},
							Raw:      []byte(root),
							Data:     nil,
							Checksum: []byte{},
						},
						"/paths.yml": {
							Info:     common.ConfigInfo{Url: mustParse("/paths.yml")},
							Raw:      []byte(path),
							Data:     nil,
							Checksum: []byte{},
						},
					},
				}
				w.providers[""] = p
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				w.Start(pool)
				ch := make(chan *openapi.Config)
				w.AddListener(func(c *common.Config) {
					require.NotNil(t, c)
					cfg := c.Data.(*openapi.Config)
					require.NotNil(t, cfg)
					ch <- cfg
				})

				p.ch <- p.files["/root.yml"]
				select {
				case c := <-ch:
					require.Equal(t, "foo", c.Paths["/users"].Value.Get.Summary)
				case <-time.After(10 * time.Second):
					require.Fail(t, "expected to get config")
				}

				path = strings.ReplaceAll(path, "foo", "bar")
				f := &common.Config{
					Info:     common.ConfigInfo{Url: mustParse("/paths.yml")},
					Raw:      []byte(path),
					Data:     nil,
					Checksum: []byte(path),
				}
				p.ch <- f

				select {
				case c := <-ch:
					require.Equal(t, "bar", c.Paths["/users"].Value.Get.Summary)
				case <-time.After(10 * time.Second):
					require.Fail(t, "expected to get config")
				}
			},
		},
	}

	common.Register("asyncapi", &asyncApi.Config{})
	common.Register("openapi", &asyncApi.Config{})

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}

type testproviderMap struct {
	files map[string]*common.Config
	ch    chan *common.Config
}

func (p *testproviderMap) Read(u *url.URL) (*common.Config, error) {
	if f, ok := p.files[u.String()]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("not found: %v", u.String())
}

func (p *testproviderMap) Start(ch chan *common.Config, _ *safe.Pool) error {
	p.ch = ch
	return nil
}
