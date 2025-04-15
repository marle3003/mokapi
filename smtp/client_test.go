package smtp

import (
	"crypto/tls"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"
	"mokapi/config/static"
	"mokapi/server/cert"
	"testing"
)

func TestClient(t *testing.T) {
	testcases := []struct {
		name      string
		tlsConfig func() *tls.Config
		test      func(t *testing.T, s *Server, c *Client)
	}{
		{
			name: "send mail",
			test: func(t *testing.T, s *Server, c *Client) {
				s.Handler = HandlerFunc(func(rw ResponseWriter, r Request) {
					ctx := ClientFromContext(r.Context())
					switch req := r.(type) {
					case *MailRequest:
						require.Equal(t, "alice@mokapi.io", req.From)
						err := rw.Write(&MailResponse{Result: Ok})
						require.NoError(t, err)
						ctx.From = req.From
					case *RcptRequest:
						require.Equal(t, "carol@mokapi.io", req.To)
						ctx.To = append(ctx.To, req.To)
						err := rw.Write(&MailResponse{Result: Ok})
						require.NoError(t, err)
					case *DataRequest:
						msg := req.Message
						require.Equal(t, "text/plain; charset=UTF-8", msg.ContentType)
						require.Equal(t, "Test Mail", msg.Subject)
						require.Equal(t, "Hello Carol", msg.Body)
					}
				})

				msg := &Message{
					Sender:  &Address{Address: "alice@mokapi.io"},
					To:      []Address{{Address: "carol@mokapi.io"}},
					Subject: "Test Mail",
					Body:    "Hello Carol",
				}
				err := c.Send(*msg.Sender, msg.To, msg)
				require.NoError(t, err)
			},
		},
		{
			name: "use TLS",
			tlsConfig: func() *tls.Config {
				store, err := cert.NewStore(&static.Config{})
				if err != nil {
					panic(err)
				}
				return &tls.Config{GetCertificate: store.GetCertificate}
			},
			test: func(t *testing.T, s *Server, c *Client) {
				s.Handler = HandlerFunc(func(rw ResponseWriter, r Request) {
					ctx := ClientFromContext(r.Context())
					switch req := r.(type) {
					case *MailRequest:
						require.Equal(t, "alice@mokapi.io", req.From)
						err := rw.Write(&MailResponse{Result: Ok})
						require.NoError(t, err)
						ctx.From = req.From
					case *RcptRequest:
						require.Equal(t, "carol@mokapi.io", req.To)
						ctx.To = append(ctx.To, req.To)
						err := rw.Write(&MailResponse{Result: Ok})
						require.NoError(t, err)
					case *DataRequest:
						msg := req.Message
						require.Equal(t, "text/plain; charset=UTF-8", msg.ContentType)
						require.Equal(t, "Test Mail", msg.Subject)
						require.Equal(t, "Hello Carol", msg.Body)
					}
				})

				msg := &Message{
					Sender:  &Address{Address: "alice@mokapi.io"},
					To:      []Address{{Address: "carol@mokapi.io"}},
					Subject: "Test Mail",
					Body:    "Hello Carol",
				}
				err := c.Send(*msg.Sender, msg.To, msg)
				require.NoError(t, err)
			},
		},
		{
			name: "send mail using quoted-printable",
			test: func(t *testing.T, s *Server, c *Client) {
				s.Handler = HandlerFunc(func(rw ResponseWriter, r Request) {
					ctx := ClientFromContext(r.Context())
					switch req := r.(type) {
					case *MailRequest:
						require.Equal(t, "alice@mokapi.io", req.From)
						err := rw.Write(&MailResponse{Result: Ok})
						require.NoError(t, err)
						ctx.From = req.From
					case *RcptRequest:
						require.Equal(t, "carol@mokapi.io", req.To)
						ctx.To = append(ctx.To, req.To)
						err := rw.Write(&MailResponse{Result: Ok})
						require.NoError(t, err)
					case *DataRequest:
						msg := req.Message
						require.Equal(t, "text/plain; charset=UTF-8", msg.ContentType)
						require.Equal(t, "Test Mail", msg.Subject)
						require.Equal(t, "Hello Carol", msg.Body)
					}
				})

				msg := &Message{
					Sender:                  &Address{Address: "alice@mokapi.io"},
					To:                      []Address{{Address: "carol@mokapi.io"}},
					Subject:                 "Test Mail",
					Body:                    "Hello=20Carol",
					ContentTransferEncoding: "quoted-printable",
				}
				err := c.Send(*msg.Sender, msg.To, msg)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			l, err := nettest.NewLocalListener("tcp")
			require.NoError(t, err)
			server := &Server{Addr: l.Addr().String()}
			if tc.tlsConfig != nil {
				server.TLSConfig = tc.tlsConfig()
			}
			defer server.Close()
			go server.Serve(l)

			client := NewClient(server.Addr)
			defer func() {
				err := client.Close()
				require.NoError(t, err)
			}()

			tc.test(t, server, client)
		})
	}
}
