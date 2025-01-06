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
		{
			name: "get Key when url is defined",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://mokapi.io")),
				}
				require.Equal(t, "61633037-3161-3264-6332-653235343933", cfg.Info.Key())
			},
		},
		{
			name: "get Key when url is not defined",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{}
				require.Equal(t, "", cfg.Info.Key())
			},
		},
		{
			name: "kernel gets itself if no nested set",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{}
				require.Equal(t, cfg.Info, *cfg.Info.Kernel())
			},
		},
		{
			name: "kernel gets nested info",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://inner.yaml"))}
				origin := cfg.Info
				outer := dynamictest.NewConfigInfo(dynamictest.WithUrl("foo.yaml"))
				dynamic.Wrap(outer, cfg)
				require.Equal(t, origin.Key(), cfg.Info.Kernel().Key())
			},
		},
		{
			name: "match url wrapped config",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://inner.yaml"))}
				outer := dynamictest.NewConfigInfo(dynamictest.WithUrl("foo.yaml"))
				dynamic.Wrap(outer, cfg)

				u, _ := url.Parse("file://inner.yaml")
				require.True(t, cfg.Info.Match(u))
			},
		},
		{
			name: "match url",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://inner.yaml"))}
				outer := dynamictest.NewConfigInfo(dynamictest.WithUrl("foo.yaml"))
				dynamic.Wrap(outer, cfg)

				u, _ := url.Parse("foo.yaml")
				require.True(t, cfg.Info.Match(u))
			},
		},
		{
			name: "match inner url",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://inner.yaml"))}
				outer := dynamictest.NewConfigInfo(dynamictest.WithUrl("foo.yaml"))
				dynamic.Wrap(outer, cfg)

				u, _ := url.Parse("file://inner.yaml")
				require.True(t, cfg.Info.Match(u))
			},
		},
		{
			name: "not matching",
			test: func(t *testing.T) {
				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://inner.yaml"))}
				outer := dynamictest.NewConfigInfo(dynamictest.WithUrl("foo.yaml"))
				dynamic.Wrap(outer, cfg)

				u, _ := url.Parse("other.yaml")
				require.False(t, cfg.Info.Match(u))
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
