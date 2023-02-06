package acceptance

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/server/ldap"
	"mokapi/server/ldap/ldaptest"
	"time"
)

type LdapSuite struct{ BaseSuite }

func (suite *LdapSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Providers.File.Directory = "./ldap"
	suite.initCmd(cfg)
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
}

func (suite *LdapSuite) TestBind() {
	r := ldaptest.NewSimpleBindRequest(0, 3, "", "")
	client := ldaptest.NewClient("127.0.0.1:8389")
	res, err := client.Send(r)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ldap.ResultSuccess, res.Body.Children[0].Value)
}
