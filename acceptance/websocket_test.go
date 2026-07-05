package acceptance

import (
	"context"
	"mokapi/config/static"
	"mokapi/try"
	"os"
	"path"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

type WebsocketSuite struct{ BaseSuite }

func (suite *WebsocketSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	cfg.ConfigFile = path.Join(wd, "mokapi.yaml")
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./websocket"}}
	cfg.Api.Search.Enabled = true
	suite.initCmd(cfg)
}

func (suite *WebsocketSuite) Test() {
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "http://localhost:22800/chat", &websocket.DialOptions{})
	require.NoError(suite.T(), err)

	err = c.Write(ctx, websocket.MessageText, []byte(`{"text": "ping"}`))
	require.NoError(suite.T(), err)

	mt, data, err := c.Read(ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), mt, websocket.MessageText)
	require.Equal(suite.T(), `{"text":"pong"}`, string(data))
}
