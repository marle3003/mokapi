package dynamic

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/readers/file"
	"mokapi/config/static"
	"net/url"
)

type ConfigWatcher struct {
	config static.Providers

	listeners []func(c *common.File)
	stop      chan bool

	readers map[string]common.Reader
}

func NewConfigWatcher(config static.Providers) *ConfigWatcher {
	return &ConfigWatcher{
		config:  config,
		stop:    make(chan bool),
		readers: make(map[string]common.Reader)}
}

func (cw *ConfigWatcher) AddListener(listener func(c *common.File)) {
	cw.listeners = append(cw.listeners, listener)
}

func (cw *ConfigWatcher) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	if r, ok := cw.readers[u.Scheme]; ok {
		return r.Read(u, opts...)
	}
	return nil, fmt.Errorf("unsupported scheme: %v", u.String())
}

func (cw *ConfigWatcher) Close() {
	cw.stop <- true
}

func (cw *ConfigWatcher) Start() error {
	update := make(chan *common.File)
	fr := file.New(cw, common.WithListener(update))
	cw.readers["file"] = fr

	if len(cw.config.File.Filename) > 0 {
		if u, err := file.ParseUrl(cw.config.File.Filename); err != nil {
			log.Errorf("parse %v: %v", cw.config.File.Filename, err)
		} else {
			_, err := fr.Read(u, common.WithListener(update))
			if err != nil {
				log.Error(err)
			}
		}
	} else if len(cw.config.File.Directory) > 0 {
		if u, err := file.ParseUrl(cw.config.File.Directory); err != nil {
			log.Errorf("parse %v: %v", cw.config.File.Directory, err)
		} else {
			err := fr.ReadDir(u)
			if err != nil {
				log.Error(err)
			}
		}
	}

	//var gw *gitWatcher
	//if len(cw.config.Git.Url) > 0 {
	//	dir, err := ioutil.TempDir("", "mokapi_git")
	//	if err != nil {
	//		return errors.Wrap(err, "unable to create temp dir for git provider")
	//
	//	} else {
	//		log.Debugf("git temp directory: %v", dir)
	//	}
	//
	//	cw.watcher.Add(dir)
	//	gw = newGitWatcher(dir, cw.config.Git)
	//	gw.Start()
	//}
	//
	//var hw *httpWatcher
	//if len(cw.config.Http.Url) > 0 {
	//	hw = newHttpWatcher(update, cw.config.Http)
	//	err := hw.Start()
	//	if err != nil {
	//		log.Errorf("unable to start http watcher: %v", err)
	//	}
	//}

	go func() {
		defer func() {
			log.Debug("closing config watcher")
			//if gw != nil {
			//	gw.close <- true
			//}
			//if hw != nil {
			//	hw.close <- true
			//}
		}()

		for {
			select {
			case <-cw.stop:
				fr.Close()
				return
			case c := <-update:
				for _, listener := range cw.listeners {
					listener(c)
				}
			}
		}
	}()

	return nil
}
