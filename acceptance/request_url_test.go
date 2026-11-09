package acceptance

import (
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RequestUrlSuite struct{ BaseSuite }

func (suite *RequestUrlSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./request-url"}}
	suite.initCmd(cfg)
}

// TestScheme verifies that request.url in an event handler reflects the protocol
// the client used: https for a TLS server, http for a plain one.
func (suite *RequestUrlSuite) TestScheme() {
	waitForScript(suite.T(), "http://localhost:18099")

	try.GetRequest(suite.T(), "https://localhost:18098/url", nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{"scheme":"https","url":"https://localhost/url"}`, body)
		}))

	try.GetRequest(suite.T(), "http://localhost:18099/url", nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{"scheme":"http","url":"http://localhost/url"}`, body)
		}))
}
