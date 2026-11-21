package acceptance

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/providers/mail"
	"mokapi/server/cert"
	"mokapi/smtp/smtptest"
	"mokapi/try"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MailSuite struct{ BaseSuite }

func (suite *MailSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	cfg.ConfigFile = path.Join(wd, "mokapi.yaml")
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./mail"}}
	cfg.Certificates.Static = []static.Certificate{
		{Cert: "./mail/mail.mokapi.local.pem"},
	}
	cfg.Api.Search.Enabled = true
	suite.initCmd(cfg)
}

func (suite *MailSuite) TestSendMail() {
	time.Sleep(2 * time.Second)

	ca := cert.DefaultRootCert()

	err := smtptest.SendMail("from@foo.bar",
		"rcipient@foo.bar",
		"smtps://localhost:8025",
		smtptest.WithSubject("Test Mail"),
		smtptest.WithBody("This is the body"),
		smtptest.WithRootCa(ca),
	)
	require.NoError(suite.T(), err)

	// test mail API
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/Mokapi%%20MailServer/mailboxes", suite.cfg.Api.Port), nil,
		try.HasStatusCode(200),
		try.HasBody(`[{"name":"rcipient@foo.bar","numMessages":1}]`),
	)
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/Mokapi%%20MailServer/mailboxes/rcipient@foo.bar", suite.cfg.Api.Port), nil,
		try.HasStatusCode(200),
		try.HasBody(`{"name":"rcipient@foo.bar","numMessages":1,"folders":["INBOX"]}`),
	)
	var messageId string
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/Mokapi%%20MailServer/mailboxes/rcipient@foo.bar/messages", suite.cfg.Api.Port), nil,
		try.HasStatusCode(200),
		try.AssertBody(func(t *testing.T, body string) {
			var v any
			err = json.Unmarshal([]byte(body), &v)
			require.NoError(suite.T(), err)
			a := v.([]any)
			m := a[0].(map[string]any)
			require.Len(t, m, 5)
			require.NotEmpty(t, m["messageId"])
			require.NotEmpty(t, m["date"])
			require.Equal(t, []any{map[string]any{"address": "from@foo.bar"}}, m["from"])
			require.Equal(t, []any{map[string]any{"address": "rcipient@foo.bar"}}, m["to"])
			require.Equal(t, "Test Mail", m["subject"])
			messageId = m["messageId"].(string)
		}),
	)
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/messages/%v", suite.cfg.Api.Port, messageId), nil,
		try.HasStatusCode(200),
		try.AssertBody(func(t *testing.T, body string) {
			var v any
			err = json.Unmarshal([]byte(body), &v)
			require.NoError(suite.T(), err)
			m := v.(map[string]any)
			require.Len(t, m, 2)
			require.Equal(t, "Mokapi MailServer", m["service"])
			m = m["data"].(map[string]any)

			require.Len(t, m, 8)
			require.Regexp(t, ".*:8025", m["server"])
			require.Equal(t, []any{map[string]any{"address": "from@foo.bar"}}, m["from"])
			require.Equal(t, []any{map[string]any{"address": "rcipient@foo.bar"}}, m["to"])
			require.NotContains(t, m, "attachments")
			require.NotContains(t, m, "sender")
			require.NotContains(t, m, "replyTo")
			require.NotContains(t, m, "cc")
			require.NotContains(t, m, "bcc")
			require.NotEmpty(t, m["messageId"])
			require.NotContains(t, m, "inReplyTo")
			require.NotEmpty(t, m["date"])
			require.Equal(t, "Test Mail", m["subject"])
			require.NotContains(t, m, "contentType")
			require.NotContains(t, m, "contentTransferEncoding")
			require.Equal(t, "This is the body", m["body"])
			require.Greater(t, m["size"], float64(0))
		}),
	)

	err = smtptest.SendMail("from@test.bar",
		"rcipient@foo.bar",
		"smtps://localhost:8025",
		smtptest.WithSubject("Test Mail"),
		smtptest.WithRootCa(ca),
	)
	require.EqualError(suite.T(), err, "550 \"[5 1 0] Rule allowSender: Sender from@test.bar does not match allow rule: .*@foo.bar\"")

	//from := "from@foo.bar"
	//password := "super_secret_password"
	//to := "recipient@foo.bar"
	//
	//msg := fmt.Sprintf("From: %v\r\n"+
	//	"To: %v\r\n"+
	//	"Subject: Test mail\r\n\r\n"+
	//	"Body\r\n", from, to)
	//
	//auth := smtp.PlainAuth("", from, password, "localhost")
	//
	//pool := x509.NewCertPool()
	//ca, err := cert.DefaultRootCert()
	//require.NoError(suite.T(), err)
	//pool.AddCert(ca)
	//tlsConfig := &tls.Config{
	//	RootCAs: pool,
	//}
	//
	//tlsDialer := tls.Dialer{
	//	NetDialer: &net.Dialer{
	//		Timeout: 30 * time.Second,
	//	},
	//	Config: tlsConfig,
	//}
	//conn, err := tlsDialer.Dial("tcp", "localhost:8025")
	//require.NoError(suite.T(), err)
	//c, err := smtp.NewClient(conn, "localhost")
	//
	//require.NoError(suite.T(), err)
	//err = c.Auth(auth)
	//require.NoError(suite.T(), err)
	//c.Mail(from)
	//c.Rcpt(to)
	//w, err := c.Data()
	//require.NoError(suite.T(), err)
	//_, err = w.Write([]byte(msg))
	//require.NoError(suite.T(), err)
	//
	//// Send actual message
	//w.Close()
	//c.Quit()
	//require.NoError(suite.T(), err)
}

func (suite *MailSuite) TestSearch() {
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=Mokapi%%20MailServer", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			assert.NoError(t, err)
			assert.NotNil(t, data)

			assert.Equal(t, float64(5), data["total"])
		}),
	)
}

func (suite *MailSuite) TestSendMail_OldFormat() {
	err := smtptest.SendMail("from@foo.bar",
		"rcipient@foo.bar",
		"smtp://localhost:8030",
		smtptest.WithSubject("Test Mail"),
	)
	require.NoError(suite.T(), err)
}

func (suite *MailSuite) TestSendMail_Multipart() {
	ca := cert.DefaultRootCert()

	err := smtptest.SendMail("from@foo.bar",
		"rcipient@foo.bar",
		"smtp://localhost:8030",
		smtptest.WithSubject("Example multipart/mixed message"),
		smtptest.WithContentType("multipart/mixed; boundary=\"simple-boundary\""),
		smtptest.WithBody(`--simple-boundary
Content-Type: text/plain; charset="UTF-8"

Hello Bob,

This is the plain text part of the email.

--simple-boundary
Content-Type: text/plain
Content-Disposition: attachment; filename="example.txt"

This is the content of the attachment.
It can be any text data.

--simple-boundary--`),
		smtptest.WithRootCa(ca),
	)
	require.NoError(suite.T(), err)

	// test mail API
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/Mokapi%%20MailServer%%20Old/mailboxes", suite.cfg.Api.Port), nil,
		try.HasStatusCode(200),
		try.HasBody(`[{"name":"rcipient@foo.bar","numMessages":1}]`),
	)
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/Mokapi%%20MailServer%%20Old/mailboxes/rcipient@foo.bar", suite.cfg.Api.Port), nil,
		try.HasStatusCode(200),
		try.HasBody(`{"name":"rcipient@foo.bar","numMessages":1,"folders":["INBOX"]}`),
	)
	var messageId string
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/Mokapi%%20MailServer%%20Old/mailboxes/rcipient@foo.bar/messages", suite.cfg.Api.Port), nil,
		try.HasStatusCode(200),
		try.AssertBody(func(t *testing.T, body string) {
			var v any
			err = json.Unmarshal([]byte(body), &v)
			assert.NoError(suite.T(), err)
			a := v.([]any)
			m := a[0].(map[string]any)
			assert.Len(t, m, 5)
			assert.NotEmpty(t, m["messageId"])
			assert.NotEmpty(t, m["date"])
			assert.Equal(t, []any{map[string]any{"address": "from@foo.bar"}}, m["from"])
			assert.Equal(t, []any{map[string]any{"address": "rcipient@foo.bar"}}, m["to"])
			assert.Equal(t, "Example multipart/mixed message", m["subject"])
			messageId = m["messageId"].(string)
		}),
	)
	try.GetRequest(suite.T(), fmt.Sprintf("http://localhost:%v/api/services/mail/messages/%v", suite.cfg.Api.Port, messageId), nil,
		try.HasStatusCode(200),
		try.AssertBody(func(t *testing.T, body string) {
			var v any
			err = json.Unmarshal([]byte(body), &v)
			assert.NoError(suite.T(), err)
			m := v.(map[string]any)
			m = m["data"].(map[string]any)
			assert.Len(t, m, 10)
			assert.Regexp(t, ".*:8030", m["server"])
			assert.Equal(t, []any{map[string]any{"address": "from@foo.bar"}}, m["from"])
			assert.Equal(t, []any{map[string]any{"address": "rcipient@foo.bar"}}, m["to"])
			assert.Equal(t, []any{
				map[string]any{
					"contentType": "text/plain",
					"name":        "example.txt",
					"size":        float64(64),
				},
			}, m["attachments"])
			assert.NotContains(t, m, "sender")
			assert.NotContains(t, m, "replyTo")
			assert.NotContains(t, m, "cc")
			assert.NotContains(t, m, "bcc")
			assert.NotEmpty(t, m["messageId"])
			assert.NotContains(t, m, "inReplyTo")
			assert.NotEmpty(t, m["date"])
			assert.Equal(t, "Example multipart/mixed message", m["subject"])
			assert.Equal(t, "text/plain; charset=\"UTF-8\"", m["contentType"])
			assert.NotContains(t, m, "contentTransferEncoding")
			assert.Equal(t, "Hello Bob,\n\nThis is the plain text part of the email.\n", m["body"])
			assert.Greater(t, m["size"], float64(0))
		}),
	)
}

func (suite *MailSuite) TestCustomServerCertificate() {
	var conn net.Conn
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		conn, err = net.Dial("tcp", "localhost:8993")
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	require.NoError(suite.T(), err)
	defer func() { _ = conn.Close() }()

	// Upgrade to TLS
	tlsConn := tls.Client(conn, &tls.Config{
		// triggers hostname verification
		ServerName:         "mail.mokapi.local",
		InsecureSkipVerify: true,
	})
	defer func() { _ = tlsConn.Close() }()

	err = tlsConn.Handshake()
	require.NoError(suite.T(), err)

	// Get server certificate
	state := tlsConn.ConnectionState()
	require.Len(suite.T(), state.PeerCertificates, 1)
	c := state.PeerCertificates[0]

	require.Equal(suite.T(), "mail.mokapi.local", c.Subject.CommonName)
}

func (suite *MailSuite) TestDynamicServerCertificate() {
	conn, err := net.Dial("tcp", "localhost:8994")
	require.NoError(suite.T(), err)
	defer func() { _ = conn.Close() }()

	// Upgrade to TLS
	tlsConn := tls.Client(conn, &tls.Config{
		// triggers hostname verification
		ServerName:         "imap.mokapi.local",
		InsecureSkipVerify: true,
	})
	defer func() { _ = tlsConn.Close() }()

	err = tlsConn.Handshake()
	require.NoError(suite.T(), err)

	// Get server certificate
	state := tlsConn.ConnectionState()
	// root CA is included
	require.Greater(suite.T(), len(state.PeerCertificates), 1)
	c := state.PeerCertificates[0]

	require.Equal(suite.T(), "imap.mokapi.local", c.Subject.CommonName)
}

// MailReloadSuite checks what a reload of the mail files changes on a running Mokapi: rules apply
// to the running server, received mails stay and the mailboxes follow the files. Every scenario
// has its own service and files, so they do not influence each other.
type MailReloadSuite struct {
	BaseSuite
	dir         string
	rulesPort   int
	mailsPort   int
	mergedPortA int
	mergedPortB int
}

// mailFile describes a mail file written by the suite. servers and body are YAML blocks.
type mailFile struct {
	title       string
	description string
	servers     string
	body        string
}

func (suite *MailReloadSuite) SetupSuite() {
	suite.rulesPort = try.GetFreePort()
	suite.mailsPort = try.GetFreePort()
	suite.mergedPortA = try.GetFreePort()
	suite.mergedPortB = try.GetFreePort()
	suite.dir = suite.T().TempDir()

	suite.writeMail("rules.yml", mailFile{title: "ReloadRules", servers: suite.smtpServer("smtp", suite.rulesPort)})
	suite.writeMail("mails.yml", mailFile{
		title:   "ReloadMails",
		servers: suite.smtpServer("smtp", suite.mailsPort),
		body: `mailboxes:
  alice@mokapi.io:
    folders:
      INBOX: {}
`,
	})
	suite.writeMail("mailboxes.yml", mailFile{title: "ReloadMailboxes", body: suite.mailboxes("alice@mokapi.io", "bob@mokapi.io")})
	suite.writeMail("merged-a.yml", mailFile{title: "ReloadMerged", servers: suite.smtpServer("a", suite.mergedPortA)})
	suite.writeMail("merged-b.yml", mailFile{title: "ReloadMerged", servers: suite.smtpServer("b", suite.mergedPortB)})

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: suite.dir}}
	suite.initCmd(cfg)
}

// smtpServer returns the servers block of a file. Servers with the same name in two files of a
// service are merged into one, so a server that is meant to be added by a file needs its own name.
func (suite *MailReloadSuite) smtpServer(name string, port int) string {
	return fmt.Sprintf("servers:\n  %s:\n    host: '127.0.0.1:%d'\n    protocol: smtp\n", name, port)
}

func (suite *MailReloadSuite) mailboxes(names ...string) string {
	s := "mailboxes:\n"
	for _, name := range names {
		s += fmt.Sprintf("  %s:\n    folders:\n      INBOX: {}\n", name)
	}
	return s
}

func (suite *MailReloadSuite) writeMail(file string, f mailFile) {
	content := fmt.Sprintf("mail: 1.0\ninfo:\n  title: %s\n  description: %q\n%s%s", f.title, f.description, f.servers, f.body)
	suite.Require().NoError(os.WriteFile(filepath.Join(suite.dir, file), []byte(content), 0o600))
}

func (suite *MailReloadSuite) waitForConfig(service string, ready func(cfg *mail.Config) bool) {
	suite.T().Helper()
	try.WaitFor(suite.T(), func() bool {
		info := suite.cmd.App.Mail.Get(service)
		if info == nil {
			return false
		}
		cfg := info.Config()
		return cfg != nil && ready(cfg)
	})
}

func (suite *MailReloadSuite) sendMail(port int, to, subject string) error {
	return smtptest.SendMail("sender@mokapi.io", to, fmt.Sprintf("smtp://127.0.0.1:%d", port), smtptest.WithSubject(subject))
}

func (suite *MailReloadSuite) serviceUrl(service, path string) string {
	return fmt.Sprintf("http://127.0.0.1:%d/api/services/mail/%s%s", suite.cfg.Api.Port, service, path)
}

func isListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// TestRuleAddedToFileAppliesToRunningServer verifies that a deny rule added to the file rejects
// the next mail on the SMTP server that is already running.
func (suite *MailReloadSuite) TestRuleAddedToFileAppliesToRunningServer() {
	t := suite.T()
	suite.waitForConfig("ReloadRules", func(*mail.Config) bool { return true })
	try.WaitFor(t, func() bool { return isListening(suite.rulesPort) })
	require.NoError(t, suite.sendMail(suite.rulesPort, "alice@mokapi.io", "spam offer"))

	suite.writeMail("rules.yml", mailFile{
		title:   "ReloadRules",
		servers: suite.smtpServer("smtp", suite.rulesPort),
		body: `rules:
  denySpam:
    subject: .*spam.*
    action: deny
`,
	})
	suite.waitForConfig("ReloadRules", func(cfg *mail.Config) bool { return len(cfg.Rules) == 1 })

	try.WaitFor(t, func() bool {
		return suite.sendMail(suite.rulesPort, "alice@mokapi.io", "spam offer") != nil
	})
	require.NoError(t, suite.sendMail(suite.rulesPort, "alice@mokapi.io", "hello"))
}

// TestMailsSurviveReload verifies that the mail in a configured INBOX and an auto-created
// mailbox are still there after the file was reloaded.
func (suite *MailReloadSuite) TestMailsSurviveReload() {
	t := suite.T()
	suite.waitForConfig("ReloadMails", func(*mail.Config) bool { return true })
	try.WaitFor(t, func() bool { return isListening(suite.mailsPort) })
	require.NoError(t, suite.sendMail(suite.mailsPort, "alice@mokapi.io", "to alice"))
	require.NoError(t, suite.sendMail(suite.mailsPort, "carol@mokapi.io", "to carol"))

	suite.writeMail("mails.yml", mailFile{
		title:       "ReloadMails",
		description: "changed",
		servers:     suite.smtpServer("smtp", suite.mailsPort),
		body:        suite.mailboxes("alice@mokapi.io"),
	})
	suite.waitForConfig("ReloadMails", func(cfg *mail.Config) bool { return cfg.Info.Description == "changed" })

	try.GetRequest(t, suite.serviceUrl("ReloadMails", "/mailboxes/alice@mokapi.io"), nil,
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"name":"alice@mokapi.io","numMessages":1,"folders":["INBOX"]}`),
	)
	try.GetRequest(t, suite.serviceUrl("ReloadMails", "/mailboxes/carol@mokapi.io"), nil,
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"name":"carol@mokapi.io","numMessages":1,"folders":["INBOX"]}`),
	)
}

// TestMailboxRemovedFromFileIsGone verifies that a mailbox that is no longer in the file
// disappears from the list and returns 404.
func (suite *MailReloadSuite) TestMailboxRemovedFromFileIsGone() {
	t := suite.T()
	suite.waitForConfig("ReloadMailboxes", func(cfg *mail.Config) bool { return len(cfg.Mailboxes) == 2 })

	suite.writeMail("mailboxes.yml", mailFile{title: "ReloadMailboxes", body: suite.mailboxes("alice@mokapi.io")})
	suite.waitForConfig("ReloadMailboxes", func(cfg *mail.Config) bool { return len(cfg.Mailboxes) == 1 })

	try.GetRequest(t, suite.serviceUrl("ReloadMailboxes", "/mailboxes/bob@mokapi.io"), nil,
		try.HasStatusCode(http.StatusNotFound))
	try.GetRequest(t, suite.serviceUrl("ReloadMailboxes", "/mailboxes"), nil,
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`[{"name":"alice@mokapi.io","numMessages":0}]`),
	)
}

// TestServerOfMergedFileStartsAndStopsWithFile verifies that a server only defined in the second
// file of a merged service runs, and stops when that file is removed.
func (suite *MailReloadSuite) TestServerOfMergedFileStartsAndStopsWithFile() {
	t := suite.T()
	try.WaitFor(t, func() bool { return isListening(suite.mergedPortA) && isListening(suite.mergedPortB) })

	suite.Require().NoError(os.Remove(filepath.Join(suite.dir, "merged-b.yml")))

	try.WaitFor(t, func() bool { return !isListening(suite.mergedPortB) })
	require.True(t, isListening(suite.mergedPortA))
}
