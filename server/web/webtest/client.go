package webtest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"mokapi/config/dynamic/openapi"
	"net/http"
)

type ResponseCondition func(r *http.Response) error

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
		err = cond(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func HasStatusCode(status int) ResponseCondition {
	return func(r *http.Response) error {
		if r.StatusCode != status {
			return fmt.Errorf("expected status code %v, got %v", status, r.StatusCode)
		}
		return nil
	}
}

func HasBody(expected string, schema *openapi.Schema) ResponseCondition {
	return func(r *http.Response) error {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("failed to read body")
		}

		body := string(b)
		if expected != body {
			return fmt.Errorf("body does not match expected value")
		}
		return nil
	}
}
