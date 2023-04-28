package acceptance

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"mokapi/acceptance/cmd"
	"mokapi/config/static"
	"mokapi/runtime/events"
	"testing"
	"time"
)

func TestAcceptance(t *testing.T) {
	suite.Run(t, new(PetStoreSuite))
	suite.Run(t, new(MailSuite))
	suite.Run(t, new(LdapSuite))
}

type BaseSuite struct {
	suite.Suite
	cmd *cmd.Cmd
	cfg *static.Config
}

func (suite *BaseSuite) initCmd(cfg *static.Config) {
	suite.cfg = cfg
	cmd, err := cmd.Start(cfg)
	require.NoError(suite.T(), err)
	suite.cmd = cmd

	// wait for server start
	time.Sleep(2 * time.Second)
}

func (suite *BaseSuite) BeforeTest(_, _ string) {
	events.SetStore(20, events.NewTraits().WithNamespace("http"))
	events.SetStore(20, events.NewTraits().WithNamespace("kafka"))
	events.SetStore(20, events.NewTraits().WithNamespace("ldap"))
}

func (suite *BaseSuite) AfterTest(_, _ string) {
	events.Reset()
}

func (suite *BaseSuite) TearDownSuite() {
	if suite.cmd != nil {
		suite.cmd.Stop()
	}
}
