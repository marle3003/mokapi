package acceptance

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"mokapi/acceptance/cmd"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"testing"
	"time"
)

func TestAcceptance(t *testing.T) {
	suite.Run(t, new(PetStoreSuite))
	suite.Run(t, new(MailSuite))
}

type BaseSuite struct {
	suite.Suite
	cmd   *cmd.Cmd
	store *openapi.Config
}

func (suite *BaseSuite) initCmd(cfg *static.Config) {
	cmd, err := cmd.Start(cfg)
	require.NoError(suite.T(), err)
	suite.cmd = cmd

	// wait for server start
	time.Sleep(time.Second)
}

func (suite *BaseSuite) TearDownSuite() {
	if suite.cmd != nil {
		suite.cmd.Stop()
	}
}
