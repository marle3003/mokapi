package try

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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
		require.Equal(t, status, tr.res.StatusCode, string(tr.GetBody()))
	}
}

func HasHeader(name, value string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		v := tr.res.Header.Get(name)
		require.Equal(t, value, v)
	}
}

func HasHeaderXor(name string, values ...string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		v := tr.res.Header.Get(name)
		matched := ""
		for _, exp := range values {
			if v == exp {
				if matched != "" {
					require.Fail(t, fmt.Sprintf("header '%v' matches at least two values: '%v' and '%v'", name, matched, exp))
				}
				matched = exp
			}
		}
		if matched == "" {
			require.Fail(t, fmt.Sprintf("header '%v' value '%v' does not match one of: %v", name, v, values))
		}
	}
}

func HasBody(expected string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		body := string(tr.GetBody())
		require.Equal(t, expected, body)
	}
}

func AssertBody(assert func(t *testing.T, body string)) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		assert(t, string(tr.GetBody()))
	}
}

func BodyContains(s string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		body := string(tr.GetBody())
		require.Contains(t, body, s)
	}
}

func BodyMatch(regexp string) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		body := string(tr.GetBody())
		require.Regexp(t, regexp, body)
	}
}

// "{\"name\":\"foo\",\"servers\":[{\"url\":\"/\",\"description\":\"\"}],\"paths\":[{\"path\":\"/foo/{bar}\",\"operations\":[{\"method\":\"get\",\"deprecated\":false,\"responses\":[{\"statusCode\":\"200\",\"description\":\"foo description\",\"contents\":[{\"type\":\"application/json\",\"schema\":{\"type\":\"object\",\"properties\":{\"loop\":{\"$ref\":\"#/components/schemas/loop\",\"type\":\"object\",\"properties\":{\"loop\":{\"$ref\":\"#/components/schemas/loop\"}}}}}}]}]}]}]
func HasBodyData(expected interface{}) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		body := tr.GetBody()
		var actual map[string]interface{}
		err := json.Unmarshal(body, &actual)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
}

func BodyContainsData(expected map[string]interface{}) ResponseCondition {
	return func(t *testing.T, tr *TestResponse) {
		body := tr.GetBody()
		var actual map[string]interface{}
		err := json.Unmarshal(body, &actual)
		require.NoError(t, err)
		for k, v := range expected {
			require.Contains(t, actual, k)
			require.Equal(t, v, actual[k])
		}
	}
}

func (tr *TestResponse) GetBody() []byte {
	if tr.body != nil {
		return tr.body
	}
	var err error
	tr.body, err = io.ReadAll(tr.res.Body)
	if err != nil {
		return []byte(err.Error())
	}
	return tr.body
}
