package acceptance

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"mokapi/acceptance/cmd"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"os"
	"testing"
	"time"
)

type BaseSuite struct {
	suite.Suite
	cmd   *cmd.Cmd
	store *openapi.Config
}

func (suite *BaseSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Providers.File.Directory = "./petstore"
	cmd, err := cmd.Start(cfg)
	require.NoError(suite.T(), err)
	suite.cmd = cmd

	suite.store = &openapi.Config{}
	b, err := os.ReadFile("./petstore/openapi.yml")
	require.NoError(suite.T(), err)
	err = yaml.Unmarshal(b, &suite.store)
	require.NoError(suite.T(), err)
	err = suite.store.Parse(&common.File{Data: suite.store}, nil)
	require.NoError(suite.T(), err)

	// wait for server start
	time.Sleep(time.Second)
}

func (suite *BaseSuite) TearDownSuite() {
	suite.cmd.Stop()
}

func TestAcceptance(t *testing.T) {
	suite.Run(t, new(PetStoreSuite))
}
