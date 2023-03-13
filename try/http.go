package try

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ResponseCondition func(t *testing.T, r *http.Response)

func GetRequest(t *testing.T, url string, headers map[string]string, conditions ...ResponseCondition) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					for _, rawCert := range rawCerts {
						c, _ := x509.ParseCertificate(rawCert)
						if c.Issuer.CommonName == "Mokapi MockServer" {
							return nil
						}
					}
					return fmt.Errorf("unknown certificate")
				}},
		}}

	res, err := client.Do(req)
	require.NoError(t, err)

	for _, cond := range conditions {
		cond(t, res)
	}
}

func Handler(t *testing.T, method string, url string, headers map[string]string, body string, handler http.Handler, conditions ...ResponseCondition) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, r)

	for _, cond := range conditions {
		cond(t, rr.Result())
	}
}

func HasStatusCode(status int) ResponseCondition {
	return func(t *testing.T, r *http.Response) {
		require.Equal(t, status, r.StatusCode)
	}
}

func HasHeader(name, value string) ResponseCondition {
	return func(t *testing.T, r *http.Response) {
		v := r.Header.Get(name)
		require.Equal(t, value, v)
	}
}

func HasBody(expected string) ResponseCondition {
	return func(t *testing.T, r *http.Response) {
		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		body := string(b)

		require.Equal(t, expected, body)
	}
}

func BodyContains(s string) ResponseCondition {
	return func(t *testing.T, r *http.Response) {
		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		body := string(b)

		require.Contains(t, body, s)
	}
}
