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
