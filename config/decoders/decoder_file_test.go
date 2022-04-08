package decoders

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDecoder_Decode(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "no filename set",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
			},
		},
		{
			name: "file does not exist",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: "test.yml"}
				err := d.Decode(map[string]string{}, s)
				require.EqualError(t, err, "open test.yml: no such file or directory")
			},
		},
		{
			name: "empty file",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "")}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
			},
		},
		{
			name: "yaml schema error",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "name: {}")}
				err := d.Decode(map[string]string{}, s)
				require.Contains(t, err.Error(), "parsing YAML file")
			},
		},
		{
			name: "file with data",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "name: foobar")}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file with data",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{}
				err := d.Decode(map[string]string{"configfile": createTempFile(t, "test.yml", "name: foobar")}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file with data",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "name: foobar")}
				err := d.Decode(map[string]string{"configfile": createTempFile(t, "test.yml", "name: barfoo")}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

func createTempFile(t *testing.T, filename, data string) string {
	path := filepath.Join(t.TempDir(), filename)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	return path
}
