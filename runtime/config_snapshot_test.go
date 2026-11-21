package runtime

import (
	"mokapi/config/dynamic"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func snapshotFile(t *testing.T, rawUrl string) *dynamic.Config {
	t.Helper()
	u, err := url.Parse(rawUrl)
	require.NoError(t, err)
	return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
}

// mergeNames returns the base names of the files it received as merge result.
func mergeNames(received *[][]*dynamic.Config) func(files []*dynamic.Config) *string {
	return func(files []*dynamic.Config) *string {
		*received = append(*received, files)
		s := ""
		for _, f := range files {
			s += f.Info.Url.String() + ";"
		}
		return &s
	}
}

func TestConfigSnapshot_Add(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config

	got := s.add(snapshotFile(t, "file:///a.yml"), mergeNames(&calls))

	require.NotNil(t, got)
	require.Equal(t, "file:///a.yml;", *got)
	require.Same(t, got, s.Config())
	require.Len(t, calls, 1)
}

func TestConfigSnapshot_FilesAreSortedByBaseName(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config

	s.add(snapshotFile(t, "file:///z/a.yml"), mergeNames(&calls))
	s.add(snapshotFile(t, "file:///a/c.yml"), mergeNames(&calls))
	got := s.add(snapshotFile(t, "file:///m/b.yml"), mergeNames(&calls))

	require.Equal(t, "file:///z/a.yml;file:///m/b.yml;file:///a/c.yml;", *got)
}

func TestConfigSnapshot_AddSameUrlReplacesFile(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config

	s.add(snapshotFile(t, "file:///a.yml"), mergeNames(&calls))
	s.add(snapshotFile(t, "file:///a.yml"), mergeNames(&calls))

	require.Len(t, s.Configs(), 1)
}

func TestConfigSnapshot_Remove(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config
	a := snapshotFile(t, "file:///a.yml")
	b := snapshotFile(t, "file:///b.yml")
	s.add(a, mergeNames(&calls))
	s.add(b, mergeNames(&calls))

	got := s.remove(a, mergeNames(&calls))

	require.Equal(t, "file:///b.yml;", *got)
	require.Same(t, got, s.Config())
	require.Equal(t, []*dynamic.Config{b}, s.Configs())
}

func TestConfigSnapshot_RemoveLastFilePublishesNilWithoutMerge(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config
	a := snapshotFile(t, "file:///a.yml")
	s.add(a, mergeNames(&calls))
	calls = nil

	got := s.remove(a, mergeNames(&calls))

	require.Nil(t, got)
	require.Nil(t, s.Config())
	require.Empty(t, s.Configs())
	require.Empty(t, calls)
}

func TestConfigSnapshot_ConfigsReturnsCopy(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config
	a := snapshotFile(t, "file:///a.yml")
	s.add(a, mergeNames(&calls))

	list := s.Configs()
	list[0] = nil
	_ = append(list, snapshotFile(t, "file:///b.yml"))

	require.Equal(t, []*dynamic.Config{a}, s.Configs())
}

func TestConfigSnapshot_ZeroValue(t *testing.T) {
	var s configSnapshot[string]

	require.Nil(t, s.Config())
	require.Empty(t, s.Configs())
}

func TestConfigSnapshot_Publish(t *testing.T) {
	var s configSnapshot[string]
	v := "published"

	s.publish(&v)

	require.Same(t, &v, s.Config())
	require.Empty(t, s.Configs())
}
