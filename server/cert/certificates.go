package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

func (store *Store) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return store.GetOrCreate(info.ServerName)
}

func (store *Store) GetOrCreate(domain string) (*tls.Certificate, error) {
	if c, ok := store.Certificates[domain]; ok {
		return c, nil
	}
	c, err := store.createTlsCertficate(domain)
	if err != nil {
		return nil, err
	}
	store.Certificates[domain] = c

	return c, nil
}

func (store *Store) createTlsCertficate(domain string) (*tls.Certificate, error) {
	cert, key, err := store.CreateCertificate(x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames:              []string{domain},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5, 6},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	})
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{cert},
		PrivateKey:  key,
	}, nil
}

func (store *Store) CreateCertificate(template x509.Certificate) (cert []byte, privKey *rsa.PrivateKey, err error) {
	privKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	cert, err = x509.CreateCertificate(rand.Reader, &template, store.CaCert, &privKey.PublicKey, store.CaKey)
	if err != nil {
		return
	}

	return
}
