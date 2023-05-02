package smtp_test

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"net/textproto"
	"testing"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name    string
		handler smtp.HandlerFunc
		test    func(t *testing.T, conn *textproto.Conn)
	}{
		{
			name: "connect",
			test: testGetGreeting,
		},
		{
			name: "say hello",
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
			},
		},
		{
			name: "send NOOP",
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				mustSendLine(t, conn, "NOOP")
				mustReadLine(t, conn, "250 [2 0 0] OK")
			},
		},
		{
			name: "send QUIT",
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				mustSendLine(t, conn, "QUIT")
				mustReadLine(t, conn, "221 [2 0 0] Bye, see you soon")
			},
		},
		{
			name: "auth successfully",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				require.IsType(t, &smtp.LoginRequest{}, req)
				rw.Write(&smtp.LoginResponse{Result: &smtp.AuthSucceeded})
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendLoginShort(t, conn)
			},
		},
		{
			name: "auth successfully long",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				require.IsType(t, &smtp.LoginRequest{}, req)
				rw.Write(&smtp.LoginResponse{Result: &smtp.AuthSucceeded})
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendLoginLong(t, conn)
			},
		},
		{
			name: "auth invalid credentials",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				require.IsType(t, &smtp.LoginRequest{}, req)
				rw.Write(&smtp.LoginResponse{Result: &smtp.InvalidAuthCredentials})
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendLoginInvalidShort(t, conn)
			},
		},
		{
			name: "MAIL TO:",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				require.IsType(t, &smtp.MailRequest{}, req)
				m := req.(*smtp.MailRequest)
				require.Equal(t, "alice@foo.bar", m.From)
				rw.Write(&smtp.MailResponse{Result: smtp.Ok})
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendMail(t, conn)
			},
		},
		{
			name: "RCPT TO:",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				switch r := req.(type) {
				case *smtp.MailRequest:
					ctx := smtp.ClientFromContext(req.Context())
					ctx.From = r.From
					rw.Write(&smtp.MailResponse{Result: smtp.Ok})
				case *smtp.RcptRequest:
					require.Equal(t, "bob@foo.bar", r.To)
					rw.Write(&smtp.RcptResponse{Result: smtp.Ok})
				default:
					t.Fatalf("unexpected request: %t", r)
				}
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendMail(t, conn)
				testSendRcpt(t, conn)
			},
		},
		{
			name: "DATA without MAIL",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				mustSendLine(t, conn, "DATA")
				mustReadLine(t, conn, "503 [5 5 1] Missing MAIL/RCPT command.")
			},
		},
		{
			name: "DATA without RCPT",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				switch r := req.(type) {
				case *smtp.MailRequest:
					ctx := smtp.ClientFromContext(req.Context())
					ctx.From = r.From
					rw.Write(&smtp.MailResponse{Result: smtp.Ok})
				}
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendMail(t, conn)
				mustSendLine(t, conn, "DATA")
				mustReadLine(t, conn, "503 [5 5 1] Missing MAIL/RCPT command.")
			},
		},
		{
			name: "DATA",
			handler: func(rw smtp.ResponseWriter, req smtp.Request) {
				ctx := smtp.ClientFromContext(req.Context())
				switch r := req.(type) {
				case *smtp.MailRequest:
					ctx.From = r.From
					rw.Write(&smtp.MailResponse{Result: smtp.Ok})
				case *smtp.RcptRequest:
					ctx.To = append(ctx.To, r.To)
					rw.Write(&smtp.RcptResponse{Result: smtp.Ok})
				case *smtp.DataRequest:
					rw.Write(&smtp.DataResponse{Result: smtp.Ok})
				}
			},
			test: func(t *testing.T, conn *textproto.Conn) {
				testGetGreeting(t, conn)
				testSendElho(t, conn)
				testSendMail(t, conn)
				testSendRcpt(t, conn)
				testSendData(t, conn)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, conn, err := smtptest.NewServer(tc.handler)
			require.NoError(t, err)
			defer server.Close()

			tpc := textproto.NewConn(conn)
			tc.test(t, tpc)
		})
	}
}

func testGetGreeting(t *testing.T, conn *textproto.Conn) {
	mustReadLine(t, conn, "220 localhost ESMTP Service Ready")
}

func testSendElho(t *testing.T, conn *textproto.Conn) {
	mustSendLine(t, conn, "EHLO localhost")
	mustReadLine(t, conn, "250-Hello localhost")
	mustReadLine(t, conn, "250 AUTH LOGIN")
}

func testSendMail(t *testing.T, conn *textproto.Conn) {
	mustSendLine(t, conn, "MAIL FROM:<alice@foo.bar>")
	mustReadLine(t, conn, "250 [2 0 0] OK")
}

func testSendRcpt(t *testing.T, conn *textproto.Conn) {
	mustSendLine(t, conn, "RCPT TO:<bob@foo.bar>")
	mustReadLine(t, conn, "250 [2 0 0] OK")
}

func testSendData(t *testing.T, conn *textproto.Conn) {
	mustSendLine(t, conn, "DATA")
	mustReadLine(t, conn, "354 [2 0 0] Send message, ending in CRLF.CRLF")
	mustSendLine(t, conn, "From: alice@foo.bar")
	mustSendLine(t, conn, "To: bob@foo.bar")
	mustSendLine(t, conn, "Date: 29 Apr 2023 10:02:36 +0200")
	mustSendLine(t, conn, "Subject: Using the new SMTP client.")
	mustSendLine(t, conn, "Content-Type: text/plain; charset=us-ascii")
	mustSendLines(t, conn, "", "")
	mustSendLine(t, conn, "Hello Bob")
	mustSendLines(t, conn, "", ".", "")
	mustReadLine(t, conn, "250 [2 0 0] OK")
}

func testSendLoginShort(t *testing.T, conn *textproto.Conn) {
	username := base64.StdEncoding.EncodeToString([]byte("foo"))
	mustSendLine(t, conn, "AUTH login %v", username)

	e := base64.StdEncoding.EncodeToString([]byte("Password:"))
	mustReadLine(t, conn, fmt.Sprintf("334 %v", e))

	password := base64.StdEncoding.EncodeToString([]byte("bar"))
	mustSendLine(t, conn, password)

	mustReadLine(t, conn, "235 [2 0 0] Authentication succeeded")
}

func testSendLoginLong(t *testing.T, conn *textproto.Conn) {
	mustSendLine(t, conn, "AUTH login")

	e := base64.StdEncoding.EncodeToString([]byte("Username:"))
	mustReadLine(t, conn, fmt.Sprintf("334 %v", e))
	username := base64.StdEncoding.EncodeToString([]byte("foo"))
	mustSendLine(t, conn, username)

	e = base64.StdEncoding.EncodeToString([]byte("Password:"))
	mustReadLine(t, conn, fmt.Sprintf("334 %v", e))

	password := base64.StdEncoding.EncodeToString([]byte("foo"))
	mustSendLine(t, conn, password)

	mustReadLine(t, conn, "235 [2 0 0] Authentication succeeded")
}

func testSendLoginInvalidShort(t *testing.T, conn *textproto.Conn) {
	e := base64.StdEncoding.EncodeToString([]byte("foo"))
	err := conn.PrintfLine("AUTH login %v", e)
	require.NoError(t, err)

	e = base64.StdEncoding.EncodeToString([]byte("Password:"))
	mustReadLine(t, conn, fmt.Sprintf("334 %v", e))

	e = base64.StdEncoding.EncodeToString([]byte("foo"))
	err = conn.PrintfLine("%v", e)
	require.NoError(t, err)

	mustReadLine(t, conn, "535 [5 7 8] Authentication credentials invalid")
}

func mustReadLine(t *testing.T, conn *textproto.Conn, expected string) {
	s, err := conn.ReadLine()
	require.NoError(t, err)
	require.Equal(t, expected, s)
}

func mustSendLine(t *testing.T, conn *textproto.Conn, format string, args ...interface{}) {
	err := conn.PrintfLine(format, args...)
	require.NoError(t, err)
}

func mustSendLines(t *testing.T, conn *textproto.Conn, lines ...string) {
	for _, line := range lines {
		err := conn.PrintfLine(line)
		require.NoError(t, err)
	}
}
