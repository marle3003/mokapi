package dynamic_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"net/url"
	"testing"
	"time"
)

func TestConfigInfo(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "url with query parameter",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://mokapi.io?query=1")),
				}
				require.Equal(t, "https://mokapi.io?query=1", cfg.Info.Path())
			},
		},
		{
			name: "opaque url (windows file path as url)",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: &url.URL{Opaque: "C:\\foo.yaml"}},
				}
				require.Equal(t, "C:\\foo.yaml", cfg.Info.Path())
			},
		},
		{
			name: "update info",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
				}
				now := time.Now().Add(-time.Second)
				checksum := []byte("foo")
				cfg.Info.Update(checksum)
				require.Equal(t, checksum, cfg.Info.Checksum)
				require.Greater(t, cfg.Info.Time, now)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
