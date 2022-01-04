package webtest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

type ResponseCondition func(r *http.Response)

func GetRequest(url string, headers map[string]string, conditions ...ResponseCondition) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}
	for _, cond := range conditions {
		cond(res)
	}
	return nil
}

func HasStatusCode(t *testing.T, status int) ResponseCondition {
	return func(r *http.Response) {
		require.Equal(t, status, r.StatusCode)
	}
}

func HasBody(t *testing.T, expected string) ResponseCondition {
	return func(r *http.Response) {
		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		body := string(b)

		require.Equal(t, expected, body)
	}
}
