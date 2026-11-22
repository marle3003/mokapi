package acceptance

import (
	"fmt"
	"io"
	"mokapi/config/static"
	"mokapi/ldap"
	"mokapi/runtime/metrics"
	"mokapi/try"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type LdapSuite struct {
	BaseSuite
	Client *ldap.Client
}

func (suite *LdapSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./ldap"}}
	cfg.Api.Search.Enabled = true
	cfg.Api.Search.InMemory = true
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

// The event is pushed before the search has results, so the search index must be
// updated when the response is known. The DN only exists in the results.
func (suite *LdapSuite) TestSearchEventIsFoundByResultInDashboardSearch() {
	_, err := suite.Client.Search(&ldap.SearchRequest{
		Scope:  ldap.ScopeWholeSubtree,
		Filter: "(objectClass=foo)",
	})
	require.NoError(suite.T(), err)

	url := fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=%%2Btype:event%%20%%2Bbob", suite.cfg.Api.Port)
	try.WaitFor(suite.T(), func() bool {
		res, err := http.Get(url)
		if err != nil {
			return false
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		return strings.Contains(string(body), "(objectClass=foo)")
	})
}

// RFC 4511 section 4.1.1: a request that cannot be decoded is answered with a
// Notice of Disconnection (unsolicited ExtendedResponse, messageID 0, protocolError)
// and the server closes the connection.
func (suite *LdapSuite) TestMalformedSearchRequestSendsNoticeOfDisconnection() {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8389", 2*time.Second)
	require.NoError(suite.T(), err)
	defer conn.Close()
	require.NoError(suite.T(), conn.SetDeadline(time.Now().Add(3*time.Second)))

	// scope must be an ENUMERATED/INTEGER but is sent as OCTET STRING
	search := ber.Encode(ber.ClassApplication, ber.TypeConstructed, 3, nil, "search request")
	search.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "baseObject"))
	search.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "wholeSubtree", "scope"))
	search.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, "derefAliases"))
	search.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 0, "sizeLimit"))
	search.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 0, "timeLimit"))
	search.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, false, "typesOnly"))
	search.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, 7, "objectClass", "present filter"))
	search.AppendChild(ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "attributes"))

	msg := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	msg.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, "Message ID"))
	msg.AppendChild(search)
	_, err = conn.Write(msg.Bytes())
	require.NoError(suite.T(), err)

	p, err := ber.ReadPacket(conn)
	require.NoError(suite.T(), err, "server should answer with a Notice of Disconnection")
	require.Len(suite.T(), p.Children, 2)
	require.Equal(suite.T(), int64(0), p.Children[0].Value, "messageID of unsolicited notification")
	res := p.Children[1]
	require.Equal(suite.T(), ber.ClassApplication, res.ClassType)
	require.Equal(suite.T(), ber.Tag(24), res.Tag, "ExtendedResponse")
	require.Equal(suite.T(), int64(ldap.ProtocolError), res.Children[0].Value, "resultCode")
	require.Equal(suite.T(), "", res.Children[1].Value, "matchedDN")
	require.NotEmpty(suite.T(), res.Children[2].Value, "diagnosticMessage")
	require.Equal(suite.T(), "1.3.6.1.4.1.1466.20036", res.Children[3].Data.String(), "responseName")

	_, err = conn.Read(make([]byte, 1))
	require.ErrorIs(suite.T(), err, io.EOF, "server should close the connection")
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
		try.BodyContains(`"filter":"(objectClass=user)"`),
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
	require.Equal(suite.T(), float64(1), suite.cmd.App.Monitor.Ldap.RequestCounter.Sum(metrics.NewQuery()))
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

// LdapReloadSuite checks what happens to the directory when LDAP files are reloaded or removed.
// The file is the source: entries written by an LDAP client are not kept.
type LdapReloadSuite struct {
	BaseSuite
	dir        string
	reloadPort int
	mergedPort int
}

func (suite *LdapReloadSuite) SetupSuite() {
	suite.reloadPort = try.GetFreePort()
	suite.mergedPort = try.GetFreePort()
	suite.dir = suite.T().TempDir()

	suite.writeFile("reload.yml", suite.directory("Reload", fmt.Sprintf("127.0.0.1:%d", suite.reloadPort), "", "cn=bob,dc=mokapi,dc=io"))
	suite.writeFile("merged-a.yml", suite.directory("Merged", fmt.Sprintf("127.0.0.1:%d", suite.mergedPort), "", "cn=amy,dc=mokapi,dc=io"))
	suite.writeFile("merged-b.yml", suite.directory("Merged", "", "", "cn=leela,dc=mokapi,dc=io"))

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: suite.dir}}
	suite.initCmd(cfg)
}

func (suite *LdapReloadSuite) directory(title, address, description, dn string) string {
	server := ""
	if address != "" {
		server = fmt.Sprintf("server:\n  address: %q\n", address)
	}
	return fmt.Sprintf(`ldap: 1.0
info:
  title: %s
  description: %q
%sentries:
  - dn: %s
    objectClass: person
`, title, description, server, dn)
}

func (suite *LdapReloadSuite) writeFile(name, content string) {
	err := os.WriteFile(filepath.Join(suite.dir, name), []byte(content), 0o600)
	suite.Require().NoError(err)
}

func (suite *LdapReloadSuite) dial(port int) *ldap.Client {
	var client *ldap.Client
	try.WaitFor(suite.T(), func() bool {
		c := ldap.NewClient(fmt.Sprintf("127.0.0.1:%d", port))
		if err := c.Dial(); err != nil {
			return false
		}
		client = c
		return true
	})
	suite.T().Cleanup(client.Close)
	return client
}

func (suite *LdapReloadSuite) hasEntry(client *ldap.Client, cn string) bool {
	res, err := client.Search(&ldap.SearchRequest{
		Scope:  ldap.ScopeWholeSubtree,
		Filter: fmt.Sprintf("(cn=%s)", cn),
	})
	suite.Require().NoError(err)
	return len(res.Results) > 0
}

func (suite *LdapReloadSuite) waitForEntry(client *ldap.Client, cn string, expected bool) {
	suite.T().Helper()
	try.WaitFor(suite.T(), func() bool {
		return suite.hasEntry(client, cn) == expected
	})
}

// TestEntryAddedByClientIsGoneAfterFileReload verifies that a reload resets the directory to
// the content of the file, also on a connection that stays open across the reload.
func (suite *LdapReloadSuite) TestEntryAddedByClientIsGoneAfterFileReload() {
	client := suite.dial(suite.reloadPort)
	suite.waitForEntry(client, "bob", true)

	res, err := client.Add(&ldap.AddRequest{
		Dn: "cn=alice,dc=mokapi,dc=io",
		Attributes: []ldap.Attribute{
			{Type: "cn", Values: []string{"alice"}},
			{Type: "objectClass", Values: []string{"person"}},
		},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(uint8(ldap.Success), res.ResultCode)
	suite.Require().True(suite.hasEntry(client, "alice"))

	suite.writeFile("reload.yml", suite.directory("Reload", fmt.Sprintf("127.0.0.1:%d", suite.reloadPort), "changed", "cn=bob,dc=mokapi,dc=io"))

	suite.waitForEntry(client, "alice", false)
	suite.True(suite.hasEntry(client, "bob"))
}

// TestEntriesOfRemovedFileAreGone verifies that removing one of two files with the same
// title removes its entries and keeps the entries of the other file.
func (suite *LdapReloadSuite) TestEntriesOfRemovedFileAreGone() {
	client := suite.dial(suite.mergedPort)
	suite.waitForEntry(client, "amy", true)
	suite.waitForEntry(client, "leela", true)

	suite.Require().NoError(os.Remove(filepath.Join(suite.dir, "merged-b.yml")))

	suite.waitForEntry(client, "leela", false)
	suite.True(suite.hasEntry(client, "amy"))
}
