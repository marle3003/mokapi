package mokapi_test

import (
	"bytes"
	"io"
	"mokapi/pkg/cmd/mokapi"
	"mokapi/schema/json/generator"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMain_SampleData(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		test func(t *testing.T, out string)
	}{
		{
			name: "generate from json",
			args: []string{"sample-data", "./test/pet.json"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `{"category":{"id":36202,"name":"rabbit"},"name":"Kevin","photoUrls":[],"tags":[{"id":36098,"name":"VertexField"},{"id":27424,"name":"Lumin"},{"name":"OpalOasis"},{"name":"LunarFlare"},{"name":"VortexEdge"}],"status":"sold"}`, out)
			},
		},
		{
			name: "generate from json using count",
			args: []string{"sample-data", "./test/pet.json", "--count", "2"},
			test: func(t *testing.T, out string) {
				items := strings.Split(out, "\n")
				require.Equal(t, `{"category":{"id":36202,"name":"rabbit"},"name":"Kevin","photoUrls":[],"tags":[{"id":36098,"name":"VertexField"},{"id":27424,"name":"Lumin"},{"name":"OpalOasis"},{"name":"LunarFlare"},{"name":"VortexEdge"}],"status":"sold"}`, items[0])
				require.Equal(t, `{"id":81873,"category":{"id":81079,"name":"canary"},"name":"Fergus","photoUrls":["https://www.primaryaggregate.com/leverage/facilitate","http://www.deputymagnetic.net/plug-and-play/deliver/proactive/back-end","https://www.districtclicks-and-mortar.org/whiteboard/web-enabled"],"tags":[{"id":55199,"name":"Swift"},{"id":38640,"name":"Unity"},{"name":"Rocket"},{"id":24044,"name":"Arctic"},{"id":95940,"name":"Amity"}],"status":"pending"}`, items[1])
			},
		},
		{
			name: "generate using shorthand count followed by inline schema",
			args: []string{"sample-data", "-n", "2", `{"type":"integer"}`},
			test: func(t *testing.T, out string) {
				require.Len(t, strings.Split(out, "\n"), 2)
			},
		},
		{
			name: "generate from json output xml",
			args: []string{"sample-data", "./test/pet.json", "--output", "xml", "--input-type", "openapi"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `<Pet><Category><id>36202</id><name>rabbit</name></Category><name>Kevin</name><photoUrl></photoUrl><tag><Tag><id>36098</id><name>VertexField</name></Tag><Tag><id>27424</id><name>Lumin</name></Tag><Tag><name>OpalOasis</name></Tag><Tag><name>LunarFlare</name></Tag><Tag><name>VortexEdge</name></Tag></tag><status>sold</status></Pet>`, out)
			},
		},
		{
			name: "generate from json output xml",
			args: []string{"sample-data", "./test/pet.json", "--output", "xml", "--input-type", "openapi", "-n", "2"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `<samples>
<Pet><Category><id>36202</id><name>rabbit</name></Category><name>Kevin</name><photoUrl></photoUrl><tag><Tag><id>36098</id><name>VertexField</name></Tag><Tag><id>27424</id><name>Lumin</name></Tag><Tag><name>OpalOasis</name></Tag><Tag><name>LunarFlare</name></Tag><Tag><name>VortexEdge</name></Tag></tag><status>sold</status></Pet>
<Pet><id>81873</id><Category><id>81079</id><name>canary</name></Category><name>Fergus</name><photoUrl><photoUrl>https://www.primaryaggregate.com/leverage/facilitate</photoUrl><photoUrl>http://www.deputymagnetic.net/plug-and-play/deliver/proactive/back-end</photoUrl><photoUrl>https://www.districtclicks-and-mortar.org/whiteboard/web-enabled</photoUrl></photoUrl><tag><Tag><id>55199</id><name>Swift</name></Tag><Tag><id>38640</id><name>Unity</name></Tag><Tag><name>Rocket</name></Tag><Tag><id>24044</id><name>Arctic</name></Tag><Tag><id>95940</id><name>Amity</name></Tag></tag><status>pending</status></Pet>
</samples>`, out)
			},
		},
		{
			name: "generate from openapi",
			args: []string{"sample-data", "../../../acceptance/petstore/openapi.yml#/paths/~1pet/put/requestBody/content/application~1json/schema"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `{"category":{"id":36202,"name":"rabbit"},"name":"Kevin","photoUrls":[],"tags":[{"id":36098,"name":"VertexField"},{"id":27424,"name":"Lumin"},{"name":"OpalOasis"},{"name":"LunarFlare"},{"name":"VortexEdge"}],"status":"sold"}`, out)
			},
		},
		{
			name: "generate using avro",
			args: []string{"sample-data", `{"type": "string"}`, "--input-type", "avro", "--output", "binary"},
			test: func(t *testing.T, out string) {
				require.Equal(t, []byte{0x1c, 0x46, 0x32, 0x63, 0x6a, 0x43, 0x68, 0x4e, 0x4c, 0x44, 0x6e, 0x6d, 0x71, 0x6b, 0x59}, []byte(out))
			},
		},
		{
			name: "generate using avro object",
			args: []string{"sample-data", `{
  "type": "record",
  "name": "LongList",
  "aliases": ["LinkedLongs"],                      
  "fields" : [
    {"name": "value", "type": "long"},            
    {"name": "next", "type": ["null", "LongList"]} 
  ]
}`, "--input-type", "avro", "--output", "binary"},
			test: func(t *testing.T, out string) {
				// json variant below
				require.Equal(t, []byte{0xb7, 0xfe, 0x47, 0x0}, []byte(out))
			},
		},
		{
			name: "generate using avro object to json",
			args: []string{"sample-data", `{
  "type": "record",
  "name": "LongList",
  "aliases": ["LinkedLongs"],                      
  "fields" : [
    {"name": "value", "type": "long"},            
    {"name": "next", "type": ["null", "LongList"]} 
  ]
}`, "--input-type", "avro"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `{"value":-589724,"next":null}`, out)
			},
		},
	}

	stdOut := os.Stdout
	stdErr := os.Stderr

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(1234567)

			reader, writer, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = writer
			os.Stderr = writer
			defer func() {
				os.Stdout = stdOut
				os.Stderr = stdErr
			}()

			cmd := mokapi.NewCmdMokapi()
			cmd.SetArgs(tc.args)

			logrus.SetOutput(io.Discard)

			err = cmd.Execute()
			require.NoError(t, err)

			_ = writer.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)
			_ = reader.Close()

			tc.test(t, buf.String())
		})
	}
}

func TestMain_SampleData_Help(t *testing.T) {
	cmd := mokapi.NewCmdMokapi()
	cmd.SetArgs([]string{"sample-data", "--help"})
	var sb strings.Builder
	cmd.SetOutput(&sb)

	require.NoError(t, cmd.Execute())

	out := sb.String()
	require.Contains(t, out, "Generates random example data from a schema definition.")
	require.Contains(t, out, "Usage:\n  mokapi sample-data [flags] [FILE|URL|SCHEMA]\n")
	require.Contains(t, out, "--count")
	require.Contains(t, out, "--input-type")
	require.Contains(t, out, "--output")
	require.Contains(t, out, "mokapi sample-data ./schema.json")
	require.NotContains(t, out, "--log-level")
}

func TestMain_Help_ListsSampleData(t *testing.T) {
	cmd := mokapi.NewCmdMokapi()
	cmd.SetArgs([]string{"--help"})
	var sb strings.Builder
	cmd.SetOutput(&sb)

	require.NoError(t, cmd.Execute())

	require.Contains(t, sb.String(), "\nCommands:\n  sample-data  Generate sample data from a schema\n")
}

func TestMain_SampleData_Stdin(t *testing.T) {
	t.Run("schema piped to stdin", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		_, err = w.WriteString(`{"type":"integer"}`)
		require.NoError(t, err)
		require.NoError(t, w.Close())
		defer func(old *os.File) { os.Stdin = old }(os.Stdin)
		os.Stdin = r

		out := executeSampleData(t, "sample-data")

		require.Regexp(t, `^-?\d+$`, out)
	})

	t.Run("empty piped stdin", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		require.NoError(t, w.Close())
		defer func(old *os.File) { os.Stdin = old }(os.Stdin)
		os.Stdin = r

		cmd := mokapi.NewCmdMokapi()
		cmd.SetArgs([]string{"sample-data"})
		logrus.SetOutput(io.Discard)

		err = cmd.Execute()

		require.EqualError(t, err, "failed to read : no input provided (expected file, URL, or stdin)")
	})

	t.Run("no input and stdin is a terminal-like char device", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("/dev/null as character device is Unix-only")
		}
		// /dev/null is a character device like a terminal, so it is not treated as piped data
		devNull, err := os.Open(os.DevNull)
		require.NoError(t, err)
		defer devNull.Close()
		defer func(old *os.File) { os.Stdin = old }(os.Stdin)
		os.Stdin = devNull

		cmd := mokapi.NewCmdMokapi()
		cmd.SetArgs([]string{"sample-data"})
		logrus.SetOutput(io.Discard)

		err = cmd.Execute()

		require.EqualError(t, err, "failed to read : no input provided (expected file, URL, or stdin)")
	})
}

// executeSampleData runs the CLI and returns everything written to stdout.
func executeSampleData(t *testing.T, args ...string) string {
	t.Helper()
	stdOut := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	defer func() { os.Stdout = stdOut }()

	cmd := mokapi.NewCmdMokapi()
	cmd.SetArgs(args)
	logrus.SetOutput(io.Discard)
	execErr := cmd.Execute()

	_ = writer.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()
	require.NoError(t, execErr)
	return buf.String()
}

func TestMain_SampleData_Yaml(t *testing.T) {
	pipeToStdin := func(t *testing.T, data string) {
		t.Helper()
		r, w, err := os.Pipe()
		require.NoError(t, err)
		_, err = w.WriteString(data)
		require.NoError(t, err)
		require.NoError(t, w.Close())
		old := os.Stdin
		t.Cleanup(func() { os.Stdin = old })
		os.Stdin = r
	}

	t.Run("json schema as argument", func(t *testing.T) {
		out := executeSampleData(t, "sample-data", `{"type":"integer"}`)

		require.Regexp(t, `^-?\d+$`, out)
	})

	t.Run("yaml jsonschema piped to stdin", func(t *testing.T) {
		pipeToStdin(t, "type: integer\n")

		out := executeSampleData(t, "sample-data")

		require.Regexp(t, `^-?\d+$`, out)
	})

	t.Run("yaml openapi schema piped to stdin", func(t *testing.T) {
		pipeToStdin(t, "type: integer\n")

		out := executeSampleData(t, "sample-data", "--input-type", "openapi")

		require.Regexp(t, `^-?\d+$`, out)
	})

	t.Run("yaml avro schema piped to stdin", func(t *testing.T) {
		pipeToStdin(t, "type: string\n")

		out := executeSampleData(t, "sample-data", "--input-type", "avro")

		require.NotEmpty(t, out)
	})

	t.Run("yaml object schema piped to stdin", func(t *testing.T) {
		pipeToStdin(t, "type: object\nproperties:\n  id:\n    type: integer\nrequired: [id]\n")

		out := executeSampleData(t, "sample-data")

		require.Regexp(t, `^\{"id":-?\d+\}$`, out)
	})

	t.Run("yaml jsonschema as argument", func(t *testing.T) {
		out := executeSampleData(t, "sample-data", "type: integer")

		require.Regexp(t, `^-?\d+$`, out)
	})

	t.Run("yaml schema file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "schema.yaml")
		require.NoError(t, os.WriteFile(file, []byte("type: object\nproperties:\n  id:\n    type: integer\nrequired: [id]\nadditionalProperties: false\n"), 0o600))

		out := executeSampleData(t, "sample-data", file)

		require.Regexp(t, `^\{"id":-?\d+\}$`, out)
	})

	t.Run("yml schema file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "schema.yml")
		require.NoError(t, os.WriteFile(file, []byte("type: object\nproperties:\n  id:\n    type: integer\nrequired: [id]\nadditionalProperties: false\n"), 0o600))

		out := executeSampleData(t, "sample-data", file)

		require.Regexp(t, `^\{"id":-?\d+\}$`, out)
	})

	t.Run("multi-line yaml argument", func(t *testing.T) {
		out := executeSampleData(t, "sample-data", "type: object\nproperties:\n  id:\n    type: integer\nrequired: [id]\nadditionalProperties: false\n")

		require.Regexp(t, `^\{"id":-?\d+\}$`, out)
	})

	t.Run("json schema file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "schema.json")
		require.NoError(t, os.WriteFile(file, []byte(`{"type":"object","properties":{"id":{"type":"integer"}},"required":["id"],"additionalProperties":false}`), 0o600))

		out := executeSampleData(t, "sample-data", file)

		require.Regexp(t, `^\{"id":-?\d+\}$`, out)
	})

	t.Run("missing file reports read error", func(t *testing.T) {
		cmd := mokapi.NewCmdMokapi()
		cmd.SetArgs([]string{"sample-data", "./missing.yaml"})
		logrus.SetOutput(io.Discard)

		err := cmd.Execute()

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read ./missing.yaml")
		require.NotContains(t, err.Error(), "schema")
	})
}
