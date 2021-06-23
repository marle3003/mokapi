package dynamic

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"hash/fnv"
	"io/ioutil"
	"mokapi/config/static"
	"net/http"
	"time"
)

type httpWatcher struct {
	update chan Config
	config static.HttpProvider

	hash  uint64
	close chan bool
}

func newHttpWatcher(update chan Config, config static.HttpProvider) *httpWatcher {
	return &httpWatcher{update: update, config: config, close: make(chan bool)}
}

func (h *httpWatcher) Start() {
	go func() {
		interval, err := time.ParseDuration(h.config.PollInterval)
		if err != nil {
			log.Errorf("unable to parse interval %q: %v", h.config.PollInterval, err.Error())
			interval = time.Second * 5
		}

		ticker := time.NewTicker(interval)
		client := &http.Client{
			//Timeout: time.Duration(p.PollTimeout),
		}

		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-h.close:
				return
			case <-ticker.C:
				func() {
					res, err := client.Get(h.config.Url)
					if err != nil {
						log.Errorf("request to %q failed: %v", h.config.Url, err.Error())
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
					if h.hash == hash.Sum64() {
						return
					}

					c := &configItem{}
					err = yaml.Unmarshal(b, c)
					if err != nil {
						log.Errorf("unable to parse %q: %v", h.config.Url, err.Error())
					}
					h.update <- c.item

					h.hash = hash.Sum64()
				}()
			}
		}
	}()
}
