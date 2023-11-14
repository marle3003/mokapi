package try

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestResponse struct {
	res  *http.Response
	body []byte
}

type ResponseCondition func(t *testing.T, r *TestResponse)

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

	tr := &TestResponse{res: res}
	for _, cond := range conditions {
		cond(t, tr)
	}
}

func Handler(t *testing.T, method string, url string, headers map[string]string, body string, handler http.Handler, conditions ...ResponseCondition) {
	r, err := http.NewRequest(method, url, strings.NewReader(body))
	require.NoError(t, err)

	for k, v := range headers {
		r.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, r)

	tr := &TestResponse{res: rr.Result()}
	for _, cond := range conditions {
		cond(t, tr)
	}
}

func HasStatusCode(status int) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		require.Equal(t, status, tr.res.StatusCode)
	}
}

func HasHeader(name, value string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		v := tr.res.Header.Get(name)
		require.Equal(t, value, v)
	}
}

func HasBody(expected string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		if tr.body == nil {
			var err error
			tr.body, err = io.ReadAll(tr.res.Body)
			require.NoError(t, err)
		}

		body := string(tr.body)

		require.Equal(t, expected, body)
	}
}

func BodyContains(s string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		if tr.body == nil {
			var err error
			tr.body, err = io.ReadAll(tr.res.Body)
			require.NoError(t, err)
		}

		body := string(tr.body)

		require.Contains(t, body, s)
	}
}
