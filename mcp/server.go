package mcp

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/version"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Service struct {
	app *runtime.App
}

func NewService(app *runtime.App) *Service {
	return &Service{app: app}
}

func NewServer(app *runtime.App) http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mokapi-mcp",
		Version: version.BuildVersion,
	}, nil)

	svc := NewService(app)
	svc.registerListApiTool(server)
	svc.registerGetSpecTool(server)

	svc.registerSendHttpRequest(server)
	svc.registerProduceKafkaMessage(server)

	svc.registerGetEvents(server)

	svc.registerGetMokapiJsAPI(server)

	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{},
	)
}

func BuildUrl(cfg static.McpServer) (*url.URL, error) {
	s := fmt.Sprintf("http://:%v%v", cfg.Port, cfg.Path)
	return url.Parse(s)
}
