package server

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
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
					files: map[string]*dynamic.Config{
						"/root.yml": {
							Info: dynamic.ConfigInfo{Url: mustParse("/root.yml"), Provider: "foo"},
							Raw:  []byte(root),
							Data: nil,
						},
						"/paths.yml": {
							Info: dynamic.ConfigInfo{Url: mustParse("/paths.yml"), Provider: "foo"},
							Raw:  []byte(path),
							Data: nil,
						},
					},
				}
				w.providers[""] = p
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				w.Start(pool)
				ch := make(chan *openapi.Config)
				w.AddListener(func(c *dynamic.Config) {
					require.NotNil(t, c)
					cfg := c.Data.(*openapi.Config)
					require.NotNil(t, cfg)
					refs := c.Refs.List()
					require.Len(t, refs, 1)
					require.Equal(t, "/paths.yml", refs[0].Info.Url.Path)
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
				f := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustParse("/paths.yml"), Checksum: []byte(path)},
					Raw:  []byte(path),
					Data: nil,
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

		{
			"parse error in child",
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
      requestBody:
        $ref: 'not_found.yaml'`

				w := NewConfigWatcher(&static.Config{})
				configPath := mustParse("file.yml")
				configPath.Scheme = "foo"
				p := &testproviderMap{
					files: map[string]*dynamic.Config{
						"/root.yml": {
							Info: dynamic.ConfigInfo{Url: mustParse("/root.yml")},
							Raw:  []byte(root),
							Data: nil,
						},
						"/paths.yml": {
							Info: dynamic.ConfigInfo{Url: mustParse("/paths.yml")},
							Raw:  []byte(path),
							Data: nil,
						},
					},
				}
				w.providers[""] = p
				pool := safe.NewPool(context.Background())
				defer pool.Stop()
				w.Start(pool)

				logrus.SetOutput(io.Discard)
				hook := logtest.NewGlobal()

				p.ch <- p.files["/root.yml"]

				time.Sleep(1 * time.Second)

				require.Equal(t, "parse error /root.yml: parsing file /root.yml: parse path '/users' failed: resolve reference 'paths.yml#/paths/users' failed: parsing file /paths.yml: parse path '/users' failed: parse operation 'GET' failed: parse request body failed: resolve reference 'not_found.yaml' failed: not found: /not_found.yaml", hook.LastEntry().Message)
			},
		},
	}

	dynamic.Register("asyncapi", &asyncApi.Config{})
	dynamic.Register("openapi", &asyncApi.Config{})

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}

type testproviderMap struct {
	files map[string]*dynamic.Config
	ch    chan *dynamic.Config
}

func (p *testproviderMap) Read(u *url.URL) (*dynamic.Config, error) {
	if f, ok := p.files[u.String()]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("not found: %v", u.String())
}

func (p *testproviderMap) Start(ch chan *dynamic.Config, _ *safe.Pool) error {
	p.ch = ch
	return nil
}
