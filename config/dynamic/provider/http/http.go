package http

import (
	"context"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"io/ioutil"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Provider struct {
	config static.HttpProvider
	client *http.Client
	files  map[string]uint64

	m    sync.RWMutex
	hash uint64
}

func New(config static.HttpProvider) *Provider {
	p := &Provider{
		config: config,
		files:  make(map[string]uint64),
	}

	transport := http.DefaultTransport.(*http.Transport)
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

	p.client = &http.Client{
		Transport: transport,
		//Timeout: time.Duration(p.PollTimeout),
	}
	return p
}

func (p *Provider) Read(u *url.URL) (*common.Config, error) {
	p.m.Lock()
	defer p.m.Unlock()

	c, _, err := p.readUrl(u)
	return c, err
}

func (p *Provider) Start(ch chan *common.Config, pool *safe.Pool) error {
	_, err := url.Parse(p.config.Url)
	if err != nil {
		return err
	}
	p.files[p.config.Url] = 0

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

	return nil
}

func (p *Provider) checkFiles(ch chan *common.Config) {
	p.m.Lock()
	defer p.m.Unlock()

	for f := range p.files {
		u, _ := url.Parse(f)
		c, changed, err := p.readUrl(u)
		if err != nil {
			log.Error(err)
		} else if changed {
			ch <- c
		}
	}
}

func (p *Provider) readUrl(u *url.URL) (c *common.Config, changed bool, err error) {
	res, err := p.client.Get(u.String())
	if err != nil {
		log.Errorf("request to %q failed: %v", p.config.Url, err)
		return
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Debugf("unable to close http response: %v", err.Error())
		}
	}()

	if res.StatusCode != http.StatusOK {
		log.Errorf("received non-ok response code: %d", res.StatusCode)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("unable to read response body: %v", err.Error())
	}

	hash := fnv.New64()
	_, err = hash.Write(b)
	if err != nil {
		log.Errorf("unable to create hash: %v", err.Error())
		return
	}

	if p.files[u.String()] == hash.Sum64() {
		return
	}
	changed = true
	p.files[u.String()] = hash.Sum64()
	c = &common.Config{
		Url:          u,
		Data:         b,
		ProviderName: "http",
	}

	return
}
