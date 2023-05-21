package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/smtp"
	"mokapi/try"
	"testing"
	"time"
)

func Test_Mail(t *testing.T) {
	testcases := []struct {
		name string
		js   string
		host common.Host
		test func(t *testing.T, v *smtp.Message, err error)
	}{
		{
			name: "simple",
			js:   "send('smtp://127.0.0.1:%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, smtp.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
				r.Equal(t, smtp.Address{Address: "bob@mokapi.io"}, v.To[0])
				r.Equal(t, "A test mail", v.Subject)
				r.Equal(t, "Hello Bob", v.Body)
				r.True(t, v.Date.After(time.Now().Add(-time.Minute*1)), "send date should be in the last minute")
			},
		},
		{
			name: "multiple from and bcc",
			js:   "send('smtp://127.0.0.1:%v', {sender: 'carol@mokapi.io', from: [{name: 'Alice', address: 'alice@mokapi.io'},'charlie@mokapi.io'], bcc: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, &smtp.Address{Address: "carol@mokapi.io"}, v.Sender)
				r.Equal(t, smtp.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
				r.Equal(t, smtp.Address{Address: "charlie@mokapi.io"}, v.From[1])
				r.Equal(t, smtp.Address{Address: "bob@mokapi.io"}, v.Bcc[0])
				r.Equal(t, "A test mail", v.Subject)
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "cc",
			js:   "send('smtp://127.0.0.1:%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, cc: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, smtp.Address{Name: "Alice", Address: "alice@mokapi.io"}, v.From[0])
				r.Equal(t, smtp.Address{Address: "bob@mokapi.io"}, v.Cc[0])
				r.Equal(t, "A test mail", v.Subject)
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "messageId",
			js:   "send('smtp://127.0.0.1:%v', {messageId: '434571BC.8070702@mokapi.io', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "434571BC.8070702@mokapi.io", v.MessageId)
			},
		},
		{
			name: "replyTo",
			js:   "send('smtp://127.0.0.1:%v', {replyTo: 'carol@mokapi.io', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, smtp.Address{Address: "carol@mokapi.io"}, v.ReplyTo[0])
			},
		},
		{
			name: "inReplyTo",
			js:   "send('smtp://127.0.0.1:%v', {inReplyTo: '434571BC.8070702@mokapi.io', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "434571BC.8070702@mokapi.io", v.InReplyTo)
			},
		},
		{
			name: "contentType",
			js:   "send('smtp://127.0.0.1:%v', {contentType: 'text/html', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "text/html", v.ContentType)
			},
		},
		{
			name: "encoding",
			js:   "send('smtp://127.0.0.1:%v', {encoding: 'quoted-printable', from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "quoted-printable", v.Encoding)
			},
		},
		{
			name: "attachment",
			js:   "send('smtp://127.0.0.1:%v', {attachments: [{content: 'hello world', filename: 'foo.txt'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/plain; charset=utf-8; name=foo.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "attachment from file",
			js:   "send('smtp://127.0.0.1:%v', {attachments: [{path: 'foo.txt'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{openFile: func(file, hint string) (string, string, error) {
				if file == "foo.txt" {
					return file, "hello world", nil
				}
				return "", "", fmt.Errorf("file not found: %v", file)
			}},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/plain; charset=utf-8; name=foo.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "attachment from file overwrite filename",
			js:   "send('smtp://127.0.0.1:%v', {attachments: [{path: 'foo.txt', filename: 'test.txt'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{openFile: func(file, hint string) (string, string, error) {
				if file == "foo.txt" {
					return file, "hello world", nil
				}
				return "", "", fmt.Errorf("file not found: %v", file)
			}},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/plain; charset=utf-8; name=test.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
		{
			name: "attachment from file overwrite content type",
			js:   "send('smtp://127.0.0.1:%v', {attachments: [{path: 'foo.txt', contentType: 'text/html'}], from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})",
			host: &testHost{openFile: func(file, hint string) (string, string, error) {
				if file == "foo.txt" {
					return file, "hello world", nil
				}
				return "", "", fmt.Errorf("file not found: %v", file)
			}},
			test: func(t *testing.T, v *smtp.Message, err error) {
				r.NoError(t, err)
				r.Len(t, v.Attachments, 1)
				r.Equal(t, "hello world", string(v.Attachments[0].Data))
				r.Equal(t, "text/html; name=foo.txt", string(v.Attachments[0].ContentType))
				r.Equal(t, "Hello Bob", v.Body)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var received *smtp.Message
			h := mail.NewHandler(&mail.Config{}, enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*common.Action {
				received = args[0].(*smtp.Message)
				return nil
			}))

			port, err := try.GetFreePort()
			r.NoError(t, err)
			server := &smtp.Server{Addr: fmt.Sprintf("127.0.0.1:%v", port), Handler: h}
			go server.ListenAndServe()
			defer server.Close()

			js := fmt.Sprintf(tc.js, port)
			s, err := New("test",
				fmt.Sprintf(`import {send} from 'mokapi/mail';
						 export default function() {
						 	%v
						}`, js),
				tc.host, static.JsConfig{})
			r.NoError(t, err)

			_, err = s.RunDefault()
			tc.test(t, received, err)
		})
	}
}
