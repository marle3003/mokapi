package acceptance

import (
	"io"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAcceptance(t *testing.T) {
	suite.Run(t, new(PetStoreSuite))
	suite.Run(t, new(KafkaSuite))
	suite.Run(t, new(KafkaMessageIdSuite))
	suite.Run(t, new(MailSuite))
	suite.Run(t, new(LdapSuite))
	suite.Run(t, new(MqttSuite))
	suite.Run(t, new(WebsocketSuite))
	suite.Run(t, new(HttpSuite))
	suite.Run(t, new(DelaySuite))
}

// waitForScript waits until the fixture script of a simulation endpoint has run. The script
// registers /sim/ready last, the OpenAPI mock alone answers it with 'not-ready'.
func waitForScript(t *testing.T, baseUrl string) {
	t.Helper()
	try.WaitFor(t, func() bool {
		res, err := http.Get(baseUrl + "/sim/ready")
		if err != nil {
			return false
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return string(b) == "ready"
	})
}

type BaseSuite struct {
	suite.Suite
	cmd *Cmd
	cfg *static.Config
}

func (suite *BaseSuite) initCmd(cfg *static.Config) {
	logrus.SetOutput(io.Discard)

	suite.cfg = cfg
	cmd, err := Start(cfg)
	require.NoError(suite.T(), err)
	suite.cmd = cmd
}

func (suite *BaseSuite) TearDownSuite() {
	if suite.cmd != nil {
		suite.cmd.Stop()
	}
}
