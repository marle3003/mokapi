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

// filesOf returns the files of a snapshot holder (a configSnapshot or an Info that embeds it).
func filesOf[T any](s interface {
	Snapshot() (*T, []*dynamic.Config)
}) []*dynamic.Config {
	_, files := s.Snapshot()
	return files
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

	require.Len(t, filesOf(&s), 1)
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
	require.Equal(t, []*dynamic.Config{b}, filesOf(&s))
}

func TestConfigSnapshot_RemoveLastFilePublishesNilWithoutMerge(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config
	a := snapshotFile(t, "file:///a.yml")
	s.add(a, mergeNames(&calls))
	calls = nil

	got := s.remove(a, mergeNames(&calls))

	require.Nil(t, got)
	cfg, files := s.Snapshot()
	require.Nil(t, cfg)
	require.Empty(t, files)
	require.Empty(t, calls)
}

func TestConfigSnapshot_SnapshotReturnsCopyOfFiles(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config
	a := snapshotFile(t, "file:///a.yml")
	s.add(a, mergeNames(&calls))

	_, list := s.Snapshot()
	list[0] = nil
	_ = append(list, snapshotFile(t, "file:///b.yml"))

	require.Equal(t, []*dynamic.Config{a}, filesOf(&s))
}

func TestConfigSnapshot_ZeroValue(t *testing.T) {
	var s configSnapshot[string]

	cfg, files := s.Snapshot()
	require.Nil(t, cfg)
	require.Empty(t, files)
}

func TestConfigSnapshot_Publish(t *testing.T) {
	var s configSnapshot[string]
	v := "published"

	s.publish(&v)

	cfg, files := s.Snapshot()
	require.Same(t, &v, cfg)
	require.Empty(t, files)
}

func TestConfigSnapshot_PublishDropsFilesOfEarlierMerge(t *testing.T) {
	var s configSnapshot[string]
	var calls [][]*dynamic.Config
	s.add(snapshotFile(t, "file:///a.yml"), mergeNames(&calls))
	v := "published"

	s.publish(&v)

	cfg, files := s.Snapshot()
	require.Same(t, &v, cfg)
	require.Empty(t, files)
}

// While a writer is inside merge (after changing the file map, before publishing the merged
// config), a reader must still get the old config together with the old files.
func TestConfigSnapshot_ConfigAndFilesAreReadTogether(t *testing.T) {
	a := snapshotFile(t, "file:///a.yml")
	b := snapshotFile(t, "file:///b.yml")

	// readPair returns the merged config (a string listing its files) and the files from one Snapshot
	readPair := func(s *configSnapshot[string]) (string, string) {
		cfg, files := s.Snapshot()
		var merged, listed string
		if cfg != nil {
			merged = *cfg
		}
		for _, f := range files {
			listed += f.Info.Url.String() + ";"
		}
		return merged, listed
	}

	t.Run("add", func(t *testing.T) {
		var s configSnapshot[string]
		var calls [][]*dynamic.Config
		s.add(a, mergeNames(&calls))

		var gotConfig, gotFiles string
		s.add(b, func(files []*dynamic.Config) *string {
			gotConfig, gotFiles = readPair(&s)
			return mergeNames(&calls)(files)
		})

		require.Equal(t, "file:///a.yml;", gotConfig)
		require.Equal(t, gotConfig, gotFiles)
		gotConfig, gotFiles = readPair(&s)
		require.Equal(t, "file:///a.yml;file:///b.yml;", gotConfig)
		require.Equal(t, gotConfig, gotFiles)
	})

	t.Run("remove", func(t *testing.T) {
		var s configSnapshot[string]
		var calls [][]*dynamic.Config
		s.add(a, mergeNames(&calls))
		s.add(b, mergeNames(&calls))
		c := snapshotFile(t, "file:///c.yml")
		s.add(c, mergeNames(&calls))

		var gotConfig, gotFiles string
		s.remove(c, func(files []*dynamic.Config) *string {
			gotConfig, gotFiles = readPair(&s)
			return mergeNames(&calls)(files)
		})

		require.Equal(t, "file:///a.yml;file:///b.yml;file:///c.yml;", gotConfig)
		require.Equal(t, gotConfig, gotFiles)
		gotConfig, gotFiles = readPair(&s)
		require.Equal(t, "file:///a.yml;file:///b.yml;", gotConfig)
		require.Equal(t, gotConfig, gotFiles)
	})
}
