package mokapi_test

import (
	"bytes"
	"io"
	"mokapi/pkg/cmd/mokapi"
	"mokapi/schema/json/generator"
	"os"
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
				require.Equal(t, `{"id":37727,"category":{"id":83580,"name":"rabbit"},"name":"Prince of Barkness","photoUrls":[],"tags":[{"id":57421,"name":"Prism"},{"id":69949,"name":"Sol"}],"status":"pending"}`, out)
			},
		},
		{
			name: "generate from json using count",
			args: []string{"sample-data", "./test/pet.json", "--count", "2"},
			test: func(t *testing.T, out string) {
				items := strings.Split(out, "\n")
				require.Equal(t, `{"id":37727,"category":{"id":83580,"name":"rabbit"},"name":"Prince of Barkness","photoUrls":[],"tags":[{"id":57421,"name":"Prism"},{"id":69949,"name":"Sol"}],"status":"pending"}`, items[0])
				require.Equal(t, `{"category":{"name":"ferret"},"name":"Demi","photoUrls":[],"tags":[{"id":56484,"name":"EchoForge"},{"id":53226,"name":"Shadow"},{"id":53241},{"id":29044,"name":"Flux"},{"id":56885,"name":"WillowSpark"}],"status":"available"}`, items[1])
			},
		},
		{
			name: "generate from json output xml",
			args: []string{"sample-data", "./test/pet.json", "--output", "xml", "--input-type", "openapi"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `<Pet><id>37727</id><Category><id>83580</id><name>rabbit</name></Category><name>Prince of Barkness</name><photoUrl></photoUrl><tag><Tag><id>57421</id><name>Prism</name></Tag><Tag><id>69949</id><name>Sol</name></Tag></tag><status>pending</status></Pet>`, out)
			},
		},
		{
			name: "generate from json output xml",
			args: []string{"sample-data", "./test/pet.json", "--output", "xml", "--input-type", "openapi", "-n", "2"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `<samples>
<Pet><id>37727</id><Category><id>83580</id><name>rabbit</name></Category><name>Prince of Barkness</name><photoUrl></photoUrl><tag><Tag><id>57421</id><name>Prism</name></Tag><Tag><id>69949</id><name>Sol</name></Tag></tag><status>pending</status></Pet>
<Pet><Category><name>ferret</name></Category><name>Demi</name><photoUrl></photoUrl><tag><Tag><id>56484</id><name>EchoForge</name></Tag><Tag><id>53226</id><name>Shadow</name></Tag><Tag><id>53241</id></Tag><Tag><id>29044</id><name>Flux</name></Tag><Tag><id>56885</id><name>WillowSpark</name></Tag></tag><status>available</status></Pet>
</samples>`, out)
			},
		},
		{
			name: "generate from openapi",
			args: []string{"sample-data", "../../../acceptance/petstore/openapi.yml#/paths/~1pet/put/requestBody/content/application~1json/schema"},
			test: func(t *testing.T, out string) {
				require.Equal(t, `{"id":37727,"category":{"id":83580,"name":"rabbit"},"name":"Prince of Barkness","photoUrls":[],"tags":[{"id":57421,"name":"Prism"},{"id":69949,"name":"Sol"}],"status":"pending"}`, out)
			},
		},
		{
			name: "generate using avro",
			args: []string{"sample-data", `{"type": "string"}`, "--input-type", "avro", "--output", "binary"},
			test: func(t *testing.T, out string) {
				require.Equal(t, []byte{0x1c, 0x46, 0x71, 0x77, 0x43, 0x72, 0x77, 0x4d, 0x66, 0x6b, 0x4f, 0x6a, 0x6f, 0x6a, 0x78}, []byte(out))
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
				require.Equal(t, []byte{0xd3, 0xb3, 0x58, 0x0}, []byte(out))
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
				require.Equal(t, `{"value":-724202,"next":null}`, out)
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
