package acceptance

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/directory"
	"mokapi/config/static"
	"mokapi/ldap"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
	"time"
)

type LdapSuite struct {
	BaseSuite
	Client *ldap.Client
}

func (suite *LdapSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Providers.File.Directory = "./ldap"
	suite.initCmd(cfg)
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	suite.Client = ldap.NewClient("127.0.0.1:8389")
	err := suite.Client.Dial()
	require.NoError(suite.T(), err)
}

func (suite *LdapSuite) TearDownSuite() {
	suite.Client.Close()
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
	require.Len(suite.T(), res.Results[0].Attributes, 2, "mail and objectClass")
}

func (suite *LdapSuite) TestLog() {
	search := &ldap.SearchRequest{
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=user)",
		Attributes: []string{"mail"},
	}
	_, err := suite.Client.Search(search)
	require.NoError(suite.T(), err)
	e := events.GetEvents(events.NewTraits().WithNamespace("ldap"))
	require.Len(suite.T(), e, 1)
	data := e[0].Data.(*directory.LdapSearchLog)
	require.Equal(suite.T(), search, data.Request)
	require.Len(suite.T(), data.Response.Results, 4)
	require.Equal(suite.T(), "Success", data.Response.Status)
}

func (suite *LdapSuite) TestMetric() {
	search := &ldap.SearchRequest{
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     "(objectClass=user)",
		Attributes: []string{"mail"},
	}
	_, err := suite.Client.Search(search)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), float64(1), suite.cmd.App.Monitor.Ldap.Search.Sum())
	q := metrics.NewQuery(metrics.ByNamespace("ldap"), metrics.ByName("search_timestamp"))
	require.Greater(suite.T(), suite.cmd.App.Monitor.Ldap.LastSearch.Value(q), float64(1))
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
