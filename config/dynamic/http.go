package dynamic

import (
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"io/ioutil"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"net/http"
	"net/url"
	"time"
)

type httpWatcher struct {
	update chan *common.File
	config static.HttpProvider
	client *http.Client

	hash  uint64
	close chan bool
}

//func (hw *httpWatcher) Read(c *Config, _ ChangeEventHandler) error {
//	hw.ReadUrl(c.Url)
//	return nil
//}

func newHttpWatcher(update chan *common.File, config static.HttpProvider) *httpWatcher {
	return &httpWatcher{update: update, config: config, close: make(chan bool)}
}

func (hw *httpWatcher) Start() error {
	u, err := url.Parse(hw.config.Url)
	if err != nil {
		return err
	}

	go func() {
		interval := time.Second * 5
		var err error
		if len(hw.config.PollInterval) > 0 {
			interval, err = time.ParseDuration(hw.config.PollInterval)
			if err != nil {
				log.Errorf("unable to parse interval %q: %v", hw.config.PollInterval, err.Error())
			}
		}

		ticker := time.NewTicker(interval)
		hw.client = &http.Client{
			//Timeout: time.Duration(p.PollTimeout),
		}

		hw.ReadUrl(u)

		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-hw.close:
				return
			case <-ticker.C:
				hw.ReadUrl(u)
			}
		}
	}()

	return nil
}

func (hw *httpWatcher) ReadUrl(url *url.URL) {
	res, err := hw.client.Get(url.String())
	if err != nil {
		log.Errorf("request to %q failed: %v", hw.config.Url, err.Error())
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
	if hw.hash == hash.Sum64() {
		return
	}

	//c := &Config{Url: url}
	//err = yaml.Unmarshal(b, c)
	//if err != nil {
	//	log.Errorf("unable to parse %q: %v", hw.config.Url, err.Error())
	//}

	//if ok := c.eventHandler(c, hw); ok {
	//	hw.update <- c
	//}

	hw.hash = hash.Sum64()
}
