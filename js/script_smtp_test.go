package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/smtp"
	"testing"
	"time"
)

func Test_Mail(t *testing.T) {
	testcases := []struct {
		name string
		js   string
		host common.Host
		test func(t *testing.T, v *mail.Mail, err error)
	}{
		{
			name: "simple",
			js:   "send('smtp://127.0.0.1:8025', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, mail.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
				r.Equal(t, mail.Address{Address: "bob@mokapi.io"}, v.To[0])
				r.Equal(t, "A test mail", v.Subject)
				r.Equal(t, "Hello Bob", v.Body)
				r.True(t, v.Date.After(time.Now().Add(-time.Minute*1)), "send date should be in the last minute")
			},
		},
		{
			name: "multiple from and bcc",
			js:   "send('smtp://127.0.0.1:8025', {sender: 'carol@mokapi.io', from: [{name: 'Alice', address: 'alice@mokapi.io'},'charlie@mokapi.io'], bcc: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, &mail.Address{Address: "carol@mokapi.io"}, v.Sender)
				r.Equal(t, mail.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
				r.Equal(t, mail.Address{Address: "charlie@mokapi.io"}, v.From[1])
				r.Equal(t, mail.Address{Address: "bob@mokapi.io"}, v.Bcc[0])
				r.Equal(t, "A test mail", v.Subject)
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "cc",
			js:   "send('smtp://127.0.0.1:8025', {from: {name: 'Alice', address: 'alice@mokapi.io'}, cc: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, mail.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
				r.Equal(t, mail.Address{Address: "bob@mokapi.io"}, v.Cc[0])
				r.Equal(t, "A test mail", v.Subject)
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "messageId",
			js:   "send('smtp://127.0.0.1:8025', {messageId: '434571BC.8070702@mokapi.io', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, "434571BC.8070702@mokapi.io", v.MessageId)
			},
		},
		{
			name: "replyTo",
			js:   "send('smtp://127.0.0.1:8025', {replyTo: 'carol@mokapi.io', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, mail.Address{Address: "carol@mokapi.io"}, v.ReplyTo[0])
			},
		},
		{
			name: "inReplyTo",
			js:   "send('smtp://127.0.0.1:8025', {inReplyTo: '434571BC.8070702@mokapi.io', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, "434571BC.8070702@mokapi.io", v.InReplyTo)
			},
		},
		{
			name: "contentType",
			js:   "send('smtp://127.0.0.1:8025', {contentType: 'text/html', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, "text/html", v.ContentType)
			},
		},
		{
			name: "encoding",
			js:   "send('smtp://127.0.0.1:8025', {encoding: 'quoted-printable', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Equal(t, "quoted-printable", v.Encoding)
			},
		},
		{
			name: "attachment",
			js:   "send('smtp://127.0.0.1:8025', {attachments: [{content: 'hello world', filename: 'foo.txt'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/plain; charset=utf-8; name=foo.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "attachment from file",
			js:   "send('smtp://127.0.0.1:8025', {attachments: [{path: 'foo.txt'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{openFile: func(file, hint string) (string, string, error) {
				if file == "foo.txt" {
					return file, "hello world", nil
				}
				return "", "", fmt.Errorf("file not found: %v", file)
			}},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/plain; charset=utf-8; name=foo.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "attachment from file overwrite filename",
			js:   "send('smtp://127.0.0.1:8025', {attachments: [{path: 'foo.txt', filename: 'test.txt'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{openFile: func(file, hint string) (string, string, error) {
				if file == "foo.txt" {
					return file, "hello world", nil
				}
				return "", "", fmt.Errorf("file not found: %v", file)
			}},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/plain; charset=utf-8; name=test.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "attachment from file overwrite content type",
			js:   "send('smtp://127.0.0.1:8025', {attachments: [{path: 'foo.txt', contentType: 'text/html'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{openFile: func(file, hint string) (string, string, error) {
				if file == "foo.txt" {
					return file, "hello world", nil
				}
				return "", "", fmt.Errorf("file not found: %v", file)
			}},
			test: func(t *testing.T, v *mail.Mail, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/html; name=foo.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var received *mail.Mail
			h := mail.NewHandler(&mail.Config{Server: "smtp://127.0.0.1:8025"}, enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*common.Action {
				received = args[0].(*mail.Mail)
				return nil
			}))

			server := &smtp.Server{Addr: "127.0.0.1:8025", Handler: h}
			go server.ListenAndServe()
			defer server.Close()

			s, err := New("test",
				fmt.Sprintf(`import {send} from 'mokapi/mail';
						 export default function() {
						 	%v
						}`, tc.js),
				tc.host, static.JsConfig{})
			r.NoError(t, err)

			_, err = s.RunDefault()
			tc.test(t, received, err)
		})
	}
}

//func TestMail_Smtps(t *testing.T) {
//	testcases := []struct {
//		name string
//		js   string
//		host common.Host
//		test func(t *testing.T, v *mail.Mail, err error)
//	}{
//		{
//			name: "simple",
//			js:   "send('smtps://127.0.0.1:8025', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
//			host: &testHost{},
//			test: func(t *testing.T, v *mail.Mail, err error) {
//				r.NoError(t, err)
//				r.Equal(t, mail.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
//				r.Equal(t, mail.Address{Address: "bob@mokapi.io"}, v.To[0])
//				r.Equal(t, "A test mail", v.Subject)
//				r.Equal(t, "Hello Bob", v.Body)
//				r.True(t, v.Date.After(time.Now().Add(-time.Minute*1)), "send date should be in the last minute")
//			},
//		},
//	}
//
//	for _, tc := range testcases {
//		tc := tc
//		t.Run(tc.name, func(t *testing.T) {
//			var received *mail.Mail
//			h := mail.NewHandler(&mail.Config{Server: "smtps://127.0.0.1:8025"}, enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*common.Action {
//				received = args[0].(*mail.Mail)
//				return nil
//			}))
//
//			store, err := cert.NewStore(&static.Config{})
//			r.NoError(t, err)
//			server := &smtp.Server{Addr: "127.0.0.1:8025", Handler: h, TLSConfig: &tls.Config{
//				GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
//					return nil, nil
//				},
//				GetCertificate: store.GetCertificate}}
//			go server.ListenAndServeTLS()
//			defer server.Close()
//
//			s, err := New("test",
//				fmt.Sprintf(`import {send} from 'mokapi/mail';
//						 export default function() {
//						 	%v
//						}`, tc.js),
//				tc.host, static.JsConfig{})
//			r.NoError(t, err)
//
//			_, err = s.RunDefault()
//			tc.test(t, received, err)
//		})
//	}
//}
