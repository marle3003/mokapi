package acceptance

import (
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SecurityEmptyOperationSuite struct{ BaseSuite }

func (suite *SecurityEmptyOperationSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./security-empty-operation"}}
	suite.initCmd(cfg)
}

// TestOperationInheritsGlobalSecurity is the contrast case: an operation without a security key
// uses the global requirement, so the script can read the credentials.
func (suite *SecurityEmptyOperationSuite) TestOperationInheritsGlobalSecurity() {
	waitForScript(suite.T(), "http://localhost:18111")

	try.GetRequest(suite.T(), "http://localhost:18111/protected",
		map[string]string{"Authorization": "Bearer 123"},
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.Equal(t, "Bearer 123", body)
		}))
}

// TestEmptyOperationSecurityRemovesGlobalSecurity verifies that "security: []" on an operation
// removes the global requirement, so no security scheme reads the credentials for it.
func (suite *SecurityEmptyOperationSuite) TestEmptyOperationSecurityRemovesGlobalSecurity() {
	waitForScript(suite.T(), "http://localhost:18111")

	try.GetRequest(suite.T(), "http://localhost:18111/public",
		map[string]string{"Authorization": "Bearer 123"},
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.Equal(t, "none", body)
		}))
}
