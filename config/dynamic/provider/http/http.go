package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type Provider struct {
	config static.HttpProvider
	client *http.Client
	files  map[string]uint64

	running bool
	m       sync.RWMutex
}

func New(config static.HttpProvider) *Provider {
	p := &Provider{
		config: config,
		files:  make(map[string]uint64),
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Errorf("failed to get system cert pool: %v", err)
		rootCAs = x509.NewCertPool()
	}
	if len(config.Ca) > 0 {
		ca, err := parseCaCert(config)
		if err != nil {
			log.Errorf("failed to use CA certification for http provider: %v", err)
		} else {
			rootCAs.AddCert(ca)
		}
	}

	transport.TLSClientConfig = &tls.Config{RootCAs: rootCAs}

	if len(config.Proxy) > 0 {
		proxy, err := url.Parse(config.Proxy)
		if err != nil {
			log.Errorf("invalid proxy url %v: %v", config.Proxy, err)
		} else {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}
	if config.TlsSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	pollTimeout := time.Second * 5
	if len(config.PollTimeout) > 0 {
		var err error
		pollTimeout, err = time.ParseDuration(config.PollTimeout)
		if err != nil {
			pollTimeout = time.Second * 5
			log.Warnf("invalid poll timeout argument '%v', using default: %v", config.PollTimeout, err)
		}
	}

	p.client = &http.Client{
		Transport: transport,
		Timeout:   pollTimeout,
	}
	return p
}

func (p *Provider) Read(u *url.URL) (*dynamic.Config, error) {
	c, _, err := p.readUrl(u)
	return c, err
}

func (p *Provider) Start(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
	if p.running {
		return nil
	}

	var err error
	for _, u := range p.config.Urls {
		if err = checkUrl(u); err != nil {
			log.Errorf("invalid url: %v", err)
			continue
		}
		p.files[u] = 0
	}

	interval := time.Second * 5
	if len(p.config.PollInterval) > 0 {
		interval, err = time.ParseDuration(p.config.PollInterval)
		if err != nil {
			return fmt.Errorf("unable to parse interval %q: %v", p.config.PollInterval, err)
		}
	}

	pool.Go(func(ctx context.Context) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		p.checkFiles(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.checkFiles(ch)
			}
		}
	})
	p.running = true

	return nil
}

func (p *Provider) checkFiles(ch chan dynamic.ConfigEvent) {
	for f, hash := range p.files {
		u, _ := url.Parse(f)
		c, changed, err := p.readUrl(u)
		if err != nil {
			if os.IsTimeout(err) {
				log.Error(fmt.Errorf("request to %v failed: request has timed out", f))
			} else {
				log.Error(fmt.Errorf("request to %v failed: %v", f, err))
			}

		} else if changed {
			e := dynamic.ConfigEvent{Config: c, Name: f}
			if hash == 0 {
				e.Event = dynamic.Create
			} else {
				e.Event = dynamic.Update
			}
			ch <- e
		}
	}
}

func (p *Provider) readUrl(u *url.URL) (c *dynamic.Config, changed bool, err error) {
	p.m.Lock()
	defer p.m.Unlock()

	var res *http.Response
	res, err = p.client.Get(u.String())
	if err != nil {
		return
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("received non-ok response code: %d", res.StatusCode)
		return
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("read response body failed: %v", err.Error())
	}

	hash := fnv.New64()
	_, err = hash.Write(b)
	if err != nil {
		log.Errorf("create hash failed: %v", err.Error())
		return
	}

	if p.files[u.String()] == hash.Sum64() {
		return
	}
	changed = true
	p.files[u.String()] = hash.Sum64()
	c = &dynamic.Config{
		Info: dynamic.ConfigInfo{
			Url:      u,
			Provider: "http",
			Time:     time.Now(),
		},
		Raw: b,
	}

	return
}

func checkUrl(s string) error {
	_, err := url.Parse(s)
	return err
}

func parseCaCert(config static.HttpProvider) (*x509.Certificate, error) {
	ca, err := config.Ca.Read("")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert from http provider config: %w", err)
	}
	cert, err := x509.ParseCertificate(ca)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA cert: %w", err)
	}
	return cert, nil
}
