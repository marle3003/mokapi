package cert_test

import (
	"crypto/rand"
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"mokapi/config/static"
	"mokapi/config/tls"
	"mokapi/server/cert"
	"mokapi/try"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	testcases := []struct {
		name string
		cfg  func() *static.Config
		test func(t *testing.T, s *cert.Store)
	}{
		{
			name: "DefaultRootCert",
			cfg:  func() *static.Config { return &static.Config{} },
			test: func(t *testing.T, s *cert.Store) {
				root := cert.DefaultRootCert()
				require.Equal(t, "CN=Mokapi MockServer,O=Mokapi,ST=Some-State,C=AU", root.Subject.String())
			},
		},
		{
			name: "default root cert from store",
			cfg:  func() *static.Config { return &static.Config{} },
			test: func(t *testing.T, s *cert.Store) {
				require.Equal(t, "CN=Mokapi MockServer,O=Mokapi,ST=Some-State,C=AU", s.CaCert.Subject.String())
			},
		},
		{
			name: "custom root certificate",
			cfg: func() *static.Config {
				rootCert, rootKey := generateCertPEM()
				return &static.Config{
					RootCaCert: tls.FileOrContent(rootCert),
					RootCaKey:  tls.FileOrContent(rootKey),
				}
			},
			test: func(t *testing.T, s *cert.Store) {
				require.Equal(t, "CN=example.test,O=Test Org", s.CaCert.Subject.String())
			},
		},
		{
			name: "get certificate at runtime",
			cfg: func() *static.Config {
				return &static.Config{}
			},
			test: func(t *testing.T, s *cert.Store) {
				r, err := s.GetCertificate(&tls2.ClientHelloInfo{ServerName: "mokapi.io"})
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Len(t, r.Certificate, 2)

				server, err := x509.ParseCertificate(r.Certificate[0])
				require.NoError(t, err)
				require.Equal(t, "CN=mokapi.io", server.Subject.String())

				ca, err := x509.ParseCertificate(r.Certificate[1])
				require.NoError(t, err)
				require.Equal(t, "CN=Mokapi MockServer,O=Mokapi,ST=Some-State,C=AU", ca.Subject.String())

				roots := x509.NewCertPool()
				roots.AddCert(cert.DefaultRootCert())
				opts := x509.VerifyOptions{
					Roots: roots,
				}

				_, err = server.Verify(opts)
				require.NoError(t, err)
			},
		},
		{
			name: "get certificate at runtime no server name in request",
			cfg: func() *static.Config {
				return &static.Config{}
			},
			test: func(t *testing.T, s *cert.Store) {
				port := try.GetFreePort()
				server := &http.Server{
					Addr:      fmt.Sprintf(":%d", port),
					TLSConfig: &tls2.Config{GetCertificate: s.GetCertificate},
					Handler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
				}

				go func() {
					_ = server.ListenAndServeTLS("", "")
				}()
				defer func() {
					_ = server.Close()
				}()

				// wait for server start
				time.Sleep(300 * time.Millisecond)

				tr := &http.Transport{
					TLSClientConfig: &tls2.Config{
						InsecureSkipVerify: true,
					},
				}

				c := http.Client{Transport: tr}
				r, err := c.Get(fmt.Sprintf("https://127.0.0.1:%d", port))
				require.NoError(t, err)
				require.NotNil(t, r)

				cer := r.TLS.PeerCertificates[0]
				require.Equal(t, []string{"localhost"}, cer.DNSNames)
				require.Equal(t, "CN=localhost", cer.Subject.String())
				var addresses []string
				for _, addr := range cer.IPAddresses {
					addresses = append(addresses, addr.String())
				}
				require.Contains(t, addresses, "127.0.0.1")

				roots := x509.NewCertPool()
				roots.AddCert(cert.DefaultRootCert())
				opts := x509.VerifyOptions{
					Roots: roots,
				}

				_, err = cer.Verify(opts)
				require.NoError(t, err)
			},
		},
		{
			name: "get certificate defined in static config",
			cfg: func() *static.Config {
				serverCert, serverKey := generateCertPEM()
				return &static.Config{
					Certificates: static.CertificateStore{
						Static: []static.Certificate{
							{
								Cert: tls.FileOrContent(serverCert),
								Key:  tls.FileOrContent(serverKey),
							},
						},
					},
				}
			},
			test: func(t *testing.T, s *cert.Store) {
				r, err := s.GetCertificate(&tls2.ClientHelloInfo{ServerName: "example.test"})
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Len(t, r.Certificate, 1)

				server, err := x509.ParseCertificate(r.Certificate[0])
				require.NoError(t, err)
				require.Equal(t, "CN=example.test,O=Test Org", server.Subject.String())
			},
		},
		{
			name: "get certificate defined in static config from file",
			cfg: func() *static.Config {
				dir := t.TempDir()

				serverCert, serverKey := generateCertPEM()
				certFile := path.Join(dir, "cert.pem")
				keyFile := path.Join(dir, "key.pem")
				if err := os.WriteFile(certFile, []byte(serverCert), 0644); err != nil {
					panic(err)
				}

				// Write private key
				if err := os.WriteFile(keyFile, []byte(serverKey), 0644); err != nil {
					panic(err)
				}

				return &static.Config{
					Certificates: static.CertificateStore{
						Static: []static.Certificate{
							{
								Cert: tls.FileOrContent(certFile),
								Key:  tls.FileOrContent(keyFile),
							},
						},
					},
				}
			},
			test: func(t *testing.T, s *cert.Store) {
				r, err := s.GetCertificate(&tls2.ClientHelloInfo{ServerName: "example.test"})
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Len(t, r.Certificate, 1)

				server, err := x509.ParseCertificate(r.Certificate[0])
				require.NoError(t, err)
				require.Equal(t, "CN=example.test,O=Test Org", server.Subject.String())
			},
		},
		{
			name: "get certificate defined in static config from one file",
			cfg: func() *static.Config {
				dir := t.TempDir()

				serverCert, serverKey := generateCertPEM()
				combinedPEM := serverCert + serverKey
				file := path.Join(dir, "cert_with_key.pem")
				if err := os.WriteFile(file, []byte(combinedPEM), 0644); err != nil {
					panic(err)
				}

				return &static.Config{
					Certificates: static.CertificateStore{
						Static: []static.Certificate{
							{
								Cert: tls.FileOrContent(file),
							},
						},
					},
				}
			},
			test: func(t *testing.T, s *cert.Store) {
				r, err := s.GetCertificate(&tls2.ClientHelloInfo{ServerName: "example.test"})
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Len(t, r.Certificate, 1)

				server, err := x509.ParseCertificate(r.Certificate[0])
				require.NoError(t, err)
				require.Equal(t, "CN=example.test,O=Test Org", server.Subject.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := cert.NewStore(tc.cfg())
			require.NoError(t, err)

			tc.test(t, s)
		})
	}
}

func generateCertPEM() (certPEM, keyPEM string) {
	// Generate a private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "example.test",
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour), // valid for 1 day

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Self-sign the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	// Encode certificate to PEM
	certPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// Encode private key to PEM
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return string(certPEMBytes), string(keyPEMBytes)
}
