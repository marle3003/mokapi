package git

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"path/filepath"
	"testing"
	"time"
)

var files = map[string]struct{}{"LICENSE": {}, "README.md": {}, "models.yml": {}, "openapi.yml": {}}

func TestGit(t *testing.T) {
	g := New(static.GitProvider{Url: "https://github.com/marle3003/mokapi-example.git"})
	p := safe.NewPool(context.Background())
	defer func() {
		p.Stop()
	}()
	ch := make(chan *common.Config)
	err := g.Start(ch, p)
	require.NoError(t, err)

	timeout := time.After(1 * time.Second)
	i := 0
Stop:
	for {
		select {
		case <-timeout:
			break Stop
		case c := <-ch:
			i++
			name := filepath.Base(c.Url.String())
			_, ok := files[name]
			assert.True(t, ok)
		}
	}
	assert.Equal(t, len(files), i)
}
