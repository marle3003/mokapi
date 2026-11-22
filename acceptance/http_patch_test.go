package acceptance

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HttpPatchSuite checks that a relative $ref in a patch file is resolved against the directory
// of the patch file, not against the directory of the base file it patches. Every scenario has
// its own API title (the patch discriminator) and port, so they do not influence each other.
type HttpPatchSuite struct {
	BaseSuite
	addsPathUrl        string
	overridesSchemaUrl string
	responseRefUrl     string
}

const patchBaseSpec = `openapi: 3.0.3
info:
  title: %[1]s
servers:
  - url: http://localhost:%[2]d
paths:
%[3]s`

const patchSpec = `openapi: 3.0.3
info:
  title: %[1]s
paths:
%[2]s`

const petSchemaFile = `type: object
properties:
  name:
    type: string
required: [name]
enum:
  - {name: Garfield}
`

const petPathWithSchemaRef = `  /pet:
    get:
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                $ref: './schemas/pet.yml'
`

// SetupSuite writes the files of all scenarios before Mokapi starts, so that the patch files are
// read together with their base files.
func (suite *HttpPatchSuite) SetupSuite() {
	baseDir := suite.T().TempDir()
	patchDir := suite.T().TempDir()
	suite.writeFile(patchDir, "schemas/pet.yml", petSchemaFile)
	suite.writeFile(patchDir, "responses/ok.yml", `description: ok
content:
  application/json:
    schema:
      $ref: '../schemas/pet.yml'
`)

	suite.addsPathUrl = suite.writeApi(baseDir, patchDir, "PatchAddsPath", "  /health: {}\n", petPathWithSchemaRef)
	suite.overridesSchemaUrl = suite.writeApi(baseDir, patchDir, "PatchOverridesSchema", `  /pet:
    get:
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: object
                enum:
                  - {name: Odie}
`, petPathWithSchemaRef)
	suite.responseRefUrl = suite.writeApi(baseDir, patchDir, "PatchResponseRef", "  /health: {}\n", `  /pet:
    get:
      responses:
        '200':
          $ref: './responses/ok.yml'
`)

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: baseDir}, {Path: patchDir}}
	suite.initCmd(cfg)
}

// writeApi writes a base file and a patch file for an API and returns the url of the mock. The
// base file sorts before the patch file, so the patch is applied on top of it.
func (suite *HttpPatchSuite) writeApi(baseDir, patchDir, title, basePaths, patchPaths string) string {
	port := try.GetFreePort()
	suite.writeFile(baseDir, title+"-a-base.yml", fmt.Sprintf(patchBaseSpec, title, port, basePaths))
	suite.writeFile(patchDir, title+"-b-patch.yml", fmt.Sprintf(patchSpec, title, patchPaths))
	return fmt.Sprintf("http://localhost:%d", port)
}

func (suite *HttpPatchSuite) writeFile(dir, name, content string) {
	path := filepath.Join(dir, name)
	suite.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o700))
	suite.Require().NoError(os.WriteFile(path, []byte(content), 0o600))
}

func (suite *HttpPatchSuite) assertPet(url string) {
	try.GetRequest(suite.T(), url+"/pet", map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{"name":"Garfield"}`, body)
		}))
}

// TestPatchAddsPathWithRelativeSchemaRef verifies that a path added by a patch file can
// reference a schema that lies next to the patch file.
func (suite *HttpPatchSuite) TestPatchAddsPathWithRelativeSchemaRef() {
	suite.assertPet(suite.addsPathUrl)
}

// TestPatchOverridesInlineSchemaWithRelativeRef verifies that a relative $ref in a patch
// replaces the inline schema of the base file.
func (suite *HttpPatchSuite) TestPatchOverridesInlineSchemaWithRelativeRef() {
	suite.assertPet(suite.overridesSchemaUrl)
}

// TestPatchUsesRelativeResponseRef verifies that a response of a patch file can be a relative
// $ref, and that the schema inside the referenced file resolves relative to that file.
func (suite *HttpPatchSuite) TestPatchUsesRelativeResponseRef() {
	suite.assertPet(suite.responseRefUrl)
}
