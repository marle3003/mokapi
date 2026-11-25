package acceptance

import (
	"mokapi/config/static"
	"mokapi/try"
	"net/http"

	"github.com/stretchr/testify/assert"
	"testing"
)

type SecurityUnknownSchemeSuite struct{ BaseSuite }

func (suite *SecurityUnknownSchemeSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./security-unknown-scheme"}}
	suite.initCmd(cfg)
}

// TestRequirementWithoutScheme verifies that a security requirement naming a scheme that is not
// declared in components.securitySchemes is skipped and the mocked response is served.
func (suite *SecurityUnknownSchemeSuite) TestRequirementWithoutScheme() {
	for _, path := range []string{"/global", "/operation"} {
		try.GetRequest(suite.T(), "http://localhost:18110"+path, nil,
			try.HasStatusCode(http.StatusOK),
			try.AssertBody(func(t *testing.T, body string) {
				assert.Equal(t, "ok", body)
			}))
	}
}
