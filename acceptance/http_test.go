package acceptance

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/try"
	"os"
	"path/filepath"
)

type HttpSuite struct {
	BaseSuite
	portA int
	portB int
}

func (suite *HttpSuite) SetupSuite() {
	suite.portA = try.GetFreePort()
	suite.portB = try.GetFreePort()

	dir := suite.T().TempDir()
	suite.writeContract(dir, "a.json", "API A", suite.portA)
	suite.writeContract(dir, "b.json", "API B", suite.portB)

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: dir}}
	suite.initCmd(cfg)
}

func (suite *HttpSuite) writeContract(dir, file, title string, port int) {
	contract := fmt.Sprintf(`{
  "openapi": "3.1.0",
  "info": { "title": %q, "version": "1" },
  "servers": [{ "url": "http://localhost:%d" }],
  "paths": { "/x": { "get": { "responses": { "200": { "description": "ok" } } } } }
}`, title, port)
	err := os.WriteFile(filepath.Join(dir, file), []byte(contract), 0o600)
	suite.Require().NoError(err)
}

// TestTwoApisOnDifferentPortsAreReachable verifies that every API listens on its own server
// port when several OpenAPI files are loaded at startup.
func (suite *HttpSuite) TestTwoApisOnDifferentPortsAreReachable() {
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%d/x", suite.portA), nil, try.HasStatusCode(200))
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%d/x", suite.portB), nil, try.HasStatusCode(200))
}
