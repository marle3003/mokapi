package acceptance

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/server/cert"
	"mokapi/smtp/smtptest"
	"mokapi/try"
	"net"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MailSuite struct{ BaseSuite }

func (suite *MailSuite) SetupSuite() {
	cfg := static.NewConfig()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	cfg.ConfigFile = path.Join(wd, "mokapi.yaml")
	cfg.Providers.File.Directories = []string{"./mail"}
	cfg.Certificates.Static = []static.Certificate{
		{Cert: "./mail/mail.mokapi.local.pem"},
	}
	cfg.Api.Search.Enabled = true
	suite.initCmd(cfg)
}

func (suite *MailSuite) TestSendMail() {
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
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer/mailboxes", nil,
		try.HasStatusCode(200),
		try.HasBody(`[{"name":"rcipient@foo.bar","numMessages":1}]`),
	)
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer/mailboxes/rcipient@foo.bar", nil,
		try.HasStatusCode(200),
		try.HasBody(`{"name":"rcipient@foo.bar","numMessages":1,"folders":["INBOX"]}`),
	)
	var messageId string
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer/mailboxes/rcipient@foo.bar/messages", nil,
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
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/messages/"+messageId, nil,
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
	require.EqualError(suite.T(), err, "550 [5 1 0] Rule allowSender: Sender from@test.bar does not match allow rule: .*@foo.bar")

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
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer%20Old/mailboxes", nil,
		try.HasStatusCode(200),
		try.HasBody(`[{"name":"rcipient@foo.bar","numMessages":1}]`),
	)
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer%20Old/mailboxes/rcipient@foo.bar", nil,
		try.HasStatusCode(200),
		try.HasBody(`{"name":"rcipient@foo.bar","numMessages":1,"folders":["INBOX"]}`),
	)
	var messageId string
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer%20Old/mailboxes/rcipient@foo.bar/messages", nil,
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
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/messages/"+messageId, nil,
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
