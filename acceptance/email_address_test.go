package acceptance

import (
	"encoding/json"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EmailAddressSuite struct{ BaseSuite }

func (suite *EmailAddressSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./email-address"}}
	suite.initCmd(cfg)
}

// TestResponsePropertyCustomerEmailAddressIsAnEmail verifies that a response property named
// customerEmailAddress holds an email and not a postal address.
func (suite *EmailAddressSuite) TestResponsePropertyCustomerEmailAddressIsAnEmail() {
	try.GetRequest(suite.T(), "http://localhost:18093/customer", nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var res map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &res))
			email, ok := res["customerEmailAddress"].(string)
			require.True(t, ok, body)
			_, err := mail.ParseAddress(email)
			assert.NoError(t, err, email)
		}))
}
