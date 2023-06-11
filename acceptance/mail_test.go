package acceptance

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/server/cert"
	"mokapi/server/smtp/smtptest"
)

type MailSuite struct{ BaseSuite }

func (suite *MailSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Providers.File.Directory = "./mail"
	suite.initCmd(cfg)
}

func (suite *MailSuite) TestSendMail() {
	ca, err := cert.DefaultRootCert()
	require.NoError(suite.T(), err)

	err = smtptest.SendMail("from@foo.bar",
		"rcipient@foo.bar",
		"smtps://localhost:8025",
		smtptest.WithSubject("Test Mail"),
		smtptest.WithRootCa(ca),
	)
	require.NoError(suite.T(), err)

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
