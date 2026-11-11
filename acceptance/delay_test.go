package acceptance

import (
	"io"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"time"
)

type DelaySuite struct{ BaseSuite }

func (suite *DelaySuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./delay"}}
	suite.initCmd(cfg)
}

// TestDelayDoesNotBlockOtherHandlers verifies that a handler awaiting delay() holds back only
// its own response. A request to another handler of the same script file is answered meanwhile.
func (suite *DelaySuite) TestDelayDoesNotBlockOtherHandlers() {
	const baseUrl = "http://localhost:18097"
	waitForScript(suite.T(), baseUrl)

	type result struct {
		status   int
		body     string
		duration time.Duration
		err      error
	}
	slow := make(chan result, 1)
	go func() {
		start := time.Now()
		res, err := http.Get(baseUrl + "/slow")
		if err != nil {
			slow <- result{err: err, duration: time.Since(start)}
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		slow <- result{status: res.StatusCode, body: string(body), duration: time.Since(start), err: err}
	}()

	// the server gives no signal that /slow is in flight
	time.Sleep(200 * time.Millisecond)

	start := time.Now()
	try.Request(suite.T(), http.MethodGet, baseUrl+"/fast", nil, nil,
		try.HasStatusCode(200), try.HasBody("fast"))
	suite.Less(time.Since(start), time.Second, "/fast must not wait for the delay of /slow")

	res := <-slow
	suite.Require().NoError(res.err)
	suite.Equal(200, res.status)
	suite.Equal("slow", res.body)
	suite.GreaterOrEqual(res.duration, 1900*time.Millisecond, "/slow must be delayed")
}
