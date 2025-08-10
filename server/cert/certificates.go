package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

func (store *Store) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return store.GetOrCreate(info)
}

func (store *Store) GetOrCreate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if c, ok := store.Certificates[info.ServerName]; ok {
		return c, nil
	}
	c, err := store.createTlsCertificate(info)
	if err != nil {
		return nil, err
	}
	store.Certificates[info.ServerName] = c

	return c, nil
}

func (store *Store) createTlsCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	serverName := info.ServerName
	if len(serverName) == 0 {
		host, _, _ := net.SplitHostPort(info.Conn.LocalAddr().String())
		ip := net.ParseIP(host)
		if ip != nil && ip.IsLoopback() {
			serverName = "localhost"
		} else {
			serverName = host
		}
	}
	ips, err := getAllIPs()
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	cert, key, err := store.CreateCertificate(x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: serverName,
		},
		DNSNames:     []string{serverName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 5, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		IPAddresses:  ips,
	})
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{cert, store.CaCert.Raw},
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

func getAllIPs() ([]net.IP, error) {
	var ips []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		// Skip down interfaces
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			if ip.IsMulticast() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
				continue
			}
			if ip.To4() != nil || ip.IsGlobalUnicast() {
				ips = append(ips, ip)
			}
		}
	}
	return ips, nil
}
