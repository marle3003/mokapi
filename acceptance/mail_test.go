package acceptance

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/server/cert"
	"mokapi/smtp/smtptest"
	"mokapi/try"
	"testing"
)

type MailSuite struct{ BaseSuite }

func (suite *MailSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Providers.File.Directories = []string{"./mail"}
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
	try.GetRequest(suite.T(), "http://localhost:8080/api/services/mail/Mokapi%20MailServer/mailboxes/rcipient@foo.bar/messages", nil,
		try.HasStatusCode(200),
		try.AssertBody(func(t *testing.T, body string) {
			var v any
			err = json.Unmarshal([]byte(body), &v)
			require.NoError(suite.T(), err)
			a := v.([]any)
			m := a[0].(map[string]any)
			require.Len(t, m, 8)
			require.Equal(t, "[::1]:8025", m["server"])
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
	require.EqualError(suite.T(), err, "550 [5 1 0] sender from@test.bar does not match allow rule: .*@foo.bar")

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

func (suite *MailSuite) TestSendMail_OldFormat() {
	err := smtptest.SendMail("from@foo.bar",
		"rcipient@foo.bar",
		"smtp://localhost:8030",
		smtptest.WithSubject("Test Mail"),
	)
	require.NoError(suite.T(), err)
}
