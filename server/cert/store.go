package cert

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"mokapi/config/static"
)

func init() {

}

type Store struct {
	Certificates map[string]*tls.Certificate
	CaCert       *x509.Certificate
	CaKey        *rsa.PrivateKey
}

func NewStore(config *static.Config) (*Store, error) {
	store := &Store{Certificates: make(map[string]*tls.Certificate)}
	var caCert, caKey []byte
	var err error

	if len(config.RootCaCert) > 0 && len(config.RootCaKey) > 0 {
		caCert, err = config.RootCaCert.Read("")
		if err != nil {
			return nil, err
		}

		caKey, err = config.RootCaKey.Read("")
		if err != nil {
			return nil, err
		}
	} else {
		caCert = defaultCaCert
		caKey = defaultCaKey
	}

	block, _ := pem.Decode(caCert)
	if block == nil {
		return nil, err
	}
	store.CaCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	block, _ = pem.Decode(caKey)
	if block == nil {
		return nil, fmt.Errorf("failed to parse key PEM")
	}
	store.CaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (store *Store) AddCertificate(domain string, certificate *tls.Certificate) {
	store.Certificates[domain] = certificate
}

var (
	defaultCaKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA6OoUDuKoI+MfK8tRb39bmT8cCbTafiPS82e8PiObPUHUP5BA
molofrPF+ZHWMiIqHpTeBZdhM5dhMDCWOr1bv2hqU1J/gD9v1Lm5rotqXObYjzFz
7vUhmI+jUazHgoGMNRi5RSJdZFEl4MIdthMeKY2yuNxWuCVokRe8eQ3SPf8SeUCQ
zwSodcqoe6DkM6EtNh5TnT2HMk+h3yTUw/5NCpkbXskvAMB+wrHX3ZOXXIyiDxAq
a2Iea9gz3yXAkLjBHTaCN0g5awUX7/twpOfMU8Q/qp3pB3YCK7sgOqwcC/jvWryG
7J38v5I4PadJWMsrANqGu9lp5HNULWVdJ5AnNwIDAQABAoIBAQC5bTnAzAPOZk/3
nqtbl8oFy+93bssP51dXPrvnwJMjhpgCbsZwAXr2fArd8JPVX8umgx/q1aSl3Rub
sOK5Ku2zCd60LRaitF5EvgOsiQOJqKK6BUXl9LPlcF02ddZz+Mz1rJQ7DOvLJKuC
LyhWPwwhStUBRTGo8uc3s+zxduZtQXKWCmIX0J6QDl8VJFmA3Z7UF6sYyoFJg8M/
degPiDM00idBfoiZGCUdIvdx2aDKblyoAaQYRRyBiQFBphwgmdHZ1I61Sgq61Sts
deylaw9MkBiQSFMAfypCJ/KSGDEx6jYGGlzkAPpKrhxBE73H5P/zZSJMOwKd+WZ2
pvzWPSKBAoGBAP0n1i9rK7N+cOFcUPixiHIIg0gBAfWfOa1/+KHCddrbfH/KjfMB
exQXucdcbRZUp1I1yY52Bnr9RO/+6AFjW6OrM1ZChNIZXqAPakq92Qo8Q/iwq6WR
LIfk9es/Hogj3E49rV+HjLiU/sYd+2v781qJrCREXuuU+cfHUO55g113AoGBAOuI
BVbBrBbkTH/+hj9G43PIvSyrIEXmiA2Q3Zavg3oaZj17j+0WMiXHPsFh05g703ni
6S2L4+3cgRtiPmVCy725k9lTTeBmGzG10v5eYwFZXWfad2RWQFF7UFB2zp4lathe
pLI5Z5JARVQmCkNGF6UC1hjwp4bPBphlIXEw4PRBAoGBAITkO6BqzucLse/riSnz
8B+Ebn4G4WNx8VItwnQP49Q+sc1XpEpzdYunpMONtkopqCgUyji5pk870st1cY8A
/GMfhPR8OMMdxDmbvf18SYoA9uF+xKxGWC0COUIDwWBQCt8bq35hZrKirFeN96TI
/weBC9eB+J4Md06zR73YFeijAoGAAsRsW6KW5QX1qCNTbHbmwTuAjX/NN0UIIDCt
idGyF8MrpFlJTZ3PJQZ8REVj+Qvq5im2V6SnHupSUgILfaEVR7tgc1M5PQO2GhVq
zzZIr4FfSZMViAZDZzGtteRPdSg3Yydpg/aMsvkyV1TDFhtCt9uarFxXe4v/Jdc6
V1wdqIECgYAKIkrpkSYimBafg5EDHZhZq2riGL9l815sFcKohfCV8B4uP+lMYztM
QvrVUlfYSAIw/I/yNIXt0RsCXNWcCgy58JqrIHB6EFOVARNqbxZFtbjOQpiP/lVK
uCYizn3bVFk0fRLTz8LUcNshBpsA2T70p9Ms5ht3xJ59465nTgItOA==
-----END RSA PRIVATE KEY-----`)
	defaultCaCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDgTCCAmmgAwIBAgIUKIH9RtX+NUkygYrctMnO/TCcS68wDQYJKoZIhvcNAQEL
BQAwTzELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxDzANBgNVBAoM
Bk1va2FwaTEaMBgGA1UEAwwRTW9rYXBpIE1vY2tTZXJ2ZXIwIBcNMjEwNzAzMTUw
MjM1WhgPMjEyMTA2MDkxNTAyMzVaME8xCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApT
b21lLVN0YXRlMQ8wDQYDVQQKDAZNb2thcGkxGjAYBgNVBAMMEU1va2FwaSBNb2Nr
U2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6OoUDuKoI+Mf
K8tRb39bmT8cCbTafiPS82e8PiObPUHUP5BAmolofrPF+ZHWMiIqHpTeBZdhM5dh
MDCWOr1bv2hqU1J/gD9v1Lm5rotqXObYjzFz7vUhmI+jUazHgoGMNRi5RSJdZFEl
4MIdthMeKY2yuNxWuCVokRe8eQ3SPf8SeUCQzwSodcqoe6DkM6EtNh5TnT2HMk+h
3yTUw/5NCpkbXskvAMB+wrHX3ZOXXIyiDxAqa2Iea9gz3yXAkLjBHTaCN0g5awUX
7/twpOfMU8Q/qp3pB3YCK7sgOqwcC/jvWryG7J38v5I4PadJWMsrANqGu9lp5HNU
LWVdJ5AnNwIDAQABo1MwUTAdBgNVHQ4EFgQUrTl9fzRNvmNqeY2mfI2aWei/cWow
HwYDVR0jBBgwFoAUrTl9fzRNvmNqeY2mfI2aWei/cWowDwYDVR0TAQH/BAUwAwEB
/zANBgkqhkiG9w0BAQsFAAOCAQEAqaE1j2mdlkz40UFxQmSYNjXaBE/xpjdMaRTd
Yo/jxeJrPDVgcALQwoE9c03NJUA6xM+dBHzJ4h1RpOFNVdrxnLs/tYgbrtp3TQ79
TezBZi05JSzYk/8OfRowXgEVd89YjjWEjyFdMxC9l5xpqfL2bt8SgdUD7Uh9UftL
yajnHbrF7jQG6gMbhXg7ANuNaESkL7lYXKhM0x7k+oDFy8aylfYdqYuzGAjwx9/x
L2K0x8YAMF0ztIPlt+LTRjkpyEieC1UxnZegpFSDYltV9HzNUAmeBpwknMeao5RI
zUMO32/zicLYY7TIBaRipfi66DpLiVrGB6twHOSkmdVKiW/Zmw==
-----END CERTIFICATE-----`)
)
