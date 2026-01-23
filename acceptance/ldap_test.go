package acceptance

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/ldap"
	"mokapi/runtime/metrics"
	"mokapi/try"
	"net/http"
	"time"

	"github.com/stretchr/testify/require"
)

type LdapSuite struct {
	BaseSuite
	Client *ldap.Client
}

func (suite *LdapSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []string{"./ldap"}
	suite.initCmd(cfg)
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	suite.Client = ldap.NewClient("127.0.0.1:8389")
	err := suite.Client.Dial()
	require.NoError(suite.T(), err)
}

func (suite *LdapSuite) TearDownSuite() {
	suite.Client.Close()
	suite.BaseSuite.TearDownSuite()
}

func (suite *LdapSuite) AfterTest(_, _ string) {
	suite.cmd.App.Monitor.Reset()
}

func (suite *LdapSuite) TestBind() {
	res, err := suite.Client.Bind("", "")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ldap.Success, res.Result)
}

func (suite *LdapSuite) TestSearch() {
	res, err := suite.Client.Search(&ldap.SearchRequest{
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=user)",
		Attributes: []string{"mail"},
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ldap.Success, res.Status)
	require.Len(suite.T(), res.Results, 4)
	require.True(suite.T(), hasResult(res.Results, "CN=farnsworthh,CN=users,DC=mokapi,DC=io"))
	require.Len(suite.T(), res.Results[0].Attributes, 1, "mail")
}

func (suite *LdapSuite) TestLog() {
	search := &ldap.SearchRequest{
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=user)",
		Attributes: []string{"mail"},
	}
	_, err := suite.Client.Search(search)
	require.NoError(suite.T(), err)

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/events?namespace=ldap", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.BodyContains(`:{"request":{"operation":"Search","baseDN":"","scope":"WholeSubtree","dereferencePolicy":0,"sizeLimit":0,"timeLimit":0,"typesOnly":false,"filter":"(objectClass=user)","attributes":["mail"],"controls":null}`),
	)
}

func (suite *LdapSuite) TestMetric() {
	search := &ldap.SearchRequest{
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=user)",
		Attributes: []string{"mail"},
	}
	_, err := suite.Client.Search(search)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), float64(1), suite.cmd.App.Monitor.Ldap.RequestCounter.Sum())
	q := metrics.NewQuery(metrics.ByNamespace("ldap"), metrics.ByName("request_timestamp"))
	require.Greater(suite.T(), suite.cmd.App.Monitor.Ldap.LastRequest.Value(q), float64(1))
}

func (suite *LdapSuite) TestJsEvent() {
	search := &ldap.SearchRequest{
		Scope:  ldap.ScopeWholeSubtree,
		Filter: "(objectClass=foo)",
	}
	res, err := suite.Client.Search(search)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ldap.Success, res.Status)
	require.Len(suite.T(), res.Results, 1)
	require.True(suite.T(), hasResult(res.Results, "CN=bob,CN=users,DC=mokapi,DC=io"))
	require.Len(suite.T(), res.Results[0].Attributes, 2, "mail and objectClass")
}

func hasResult(results []ldap.SearchResult, dn string) bool {
	for _, r := range results {
		if r.Dn == dn {
			return true
		}
	}
	return false
}
