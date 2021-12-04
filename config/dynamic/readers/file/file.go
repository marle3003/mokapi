package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/script"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Reader struct {
	files map[string]*common.File

	watcher *fsnotify.Watcher
	close   chan bool
	lock    sync.RWMutex
	reader  common.Reader
	opts    []common.FileOptions

	readFileFunc func(string) ([]byte, error)
}

func New(reader common.Reader, opts ...common.FileOptions) *Reader {
	r := &Reader{
		files:        make(map[string]*common.File),
		reader:       reader,
		opts:         opts,
		readFileFunc: ioutil.ReadFile,
	}
	r.Start()
	return r
}

func (fr *Reader) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	name := fr.name(u)

	fr.lock.RLock()
	file, ok := fr.files[name]
	fr.lock.RUnlock()

	if !ok {
		return fr.add(u, opts...)
	} else {
		for _, opt := range opts {
			opt(file)
		}
	}

	return file, nil
}

func (fr *Reader) add(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	fr.lock.Lock()
	name := fr.name(u)

	file, ok := fr.files[name]
	if !ok {
		file = &common.File{Url: u}

		for _, opt := range opts {
			opt(file)
		}

		if err := fr.read(file); err != nil {
			fr.lock.Unlock()
			return nil, err
		} else if file.Data == nil {
			fr.lock.Unlock()
			return nil, common.UnknownFile
		}
		fr.files[name] = file
		fr.lock.Unlock()

		if p, ok := file.Data.(common.Parser); ok {
			err := p.Parse(file, fr.reader)
			if err != nil {
				return nil, err
			}
		}
	} else {
		fr.lock.Unlock()
		for _, opt := range opts {
			opt(file)
		}
	}

	return file, nil
}

func (fr *Reader) ReadDir(u *url.URL) error {
	name := fr.name(u)

	walkDir := func(path string, fi os.FileInfo, _ error) error {
		if fi.Mode().IsDir() {
			if skipPath(path) {
				return filepath.SkipDir
			}
			return fr.watcher.Add(path)
		} else if isValidConfigFile(path) {
			u, err := ParseUrl(path)
			if err != nil {
				return fmt.Errorf("unable to parse %v: %v", path, err)
			}
			go func() {
				if f, err := fr.Read(u, fr.opts...); err != nil {
					if err != common.UnknownFile {
						log.Error(err)
					}
				} else {
					f.Changed()
				}
			}()
		}

		return nil
	}

	return filepath.Walk(name, walkDir)
}

func (fr *Reader) Start() {
	var err error
	if fr.watcher, err = fsnotify.NewWatcher(); err != nil {
		log.Error("error creating file watcher", err)
		return
	}

	ticker := time.NewTicker(time.Second)
	var events []fsnotify.Event

	go func() {
		defer func() {
			log.Info("closing file watcher. Restart is required...")
			ticker.Stop()
			err := fr.watcher.Close()
			if err != nil {
				log.Error("unable to close file watcher")
			}
		}()

		for {
			select {
			case <-fr.close:
				return
			case evt := <-fr.watcher.Events:
				// temporary files ends with '~' in name
				if len(evt.Name) > 0 && !strings.HasSuffix(evt.Name, "~") {
					events = append(events, evt)
				}
			case <-ticker.C:
				m := make(map[string]bool)
				for _, evt := range events {
					if _, ok := m[evt.Name]; ok {
						continue
					}
					m[evt.Name] = true

					if b, err := isDir(evt.Name); err != nil {
						log.Errorf("unable to read event from %v: %v", evt.Name, err)
					} else if b && !skipPath(evt.Name) {
						if err := fr.watcher.Add(evt.Name); err != nil {
							log.Error(err)
						}
					} else if !isValidConfigFile(evt.Name) {
						continue
					}

					log.Debugf("item change event received " + evt.Name)

					f, ok := fr.files[evt.Name]
					if !ok {
						u, _ := ParseUrl(evt.Name)
						if f, err := fr.Read(u, fr.opts...); err != nil {
							log.Errorf("unable to read %v: %v", evt.Name, err.Error())
							continue
						} else {
							f.Changed()
						}
					} else {
						if err := fr.read(f); err != nil {
							log.Errorf("unable to read %v: %v", evt.Name, err.Error())
							continue
						}
						if p, ok := f.Data.(common.Parser); ok {
							err := p.Parse(f, fr.reader)
							if err != nil {
								log.Errorf("parser error %v: %v", evt.Name, err)
								continue
							}
						}
					}

					f.Changed()
				}

				events = nil
			}
		}
	}()
}

func (fr *Reader) Close() {
	fr.close <- true
}

func (fr *Reader) read(file *common.File) error {
	name := fr.name(file.Url)

	data, err := fr.readFileFunc(name)
	if err != nil {
		return err
	}

	if filepath.Ext(name) == ".tmpl" {
		content := string(data)

		funcMap := sprig.TxtFuncMap()
		funcMap["extractUsername"] = extractUsername
		tmpl := template.New(name).Funcs(funcMap)

		tmpl, err = tmpl.Parse(content)
		if err != nil {
			return err
		}

		var buffer bytes.Buffer
		err = tmpl.Execute(&buffer, false)
		if err != nil {
			return err
		}

		data = buffer.Bytes()
	}

	return fr.parseConfig(name, data, file)
}

func (fr *Reader) name(u *url.URL) string {
	pathOnFs := path.Clean(u.String()[len(u.Scheme)+len(":/"):])
	if len(u.Fragment) > 0 {
		pathOnFs = pathOnFs[:len(pathOnFs)-len(u.Fragment)-1] // -1 for #
	}
	pathOnFs = strings.TrimPrefix(pathOnFs, "/")
	pathOnFs = filepath.FromSlash(pathOnFs)
	return pathOnFs
}

func (fr *Reader) parseConfig(filename string, data []byte, file *common.File) error {
	if file.AsPlainText {
		file.Data = string(data)
		return nil
	}

	switch filepath.Ext(filename) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(data, file)
		if err != nil {
			return errors.Wrapf(err, "parsing yaml file %s", filename)
		}
		return nil
	case ".json":
		err := json.Unmarshal(data, file)
		if err != nil {
			return errors.Wrapf(err, "parsing json file %s", filename)
		}
		return nil
	case ".tmpl":
		filename = filename[0 : len(filename)-len(filepath.Ext(filename))]
		return fr.parseConfig(filename, data, file)
	case ".lua":
		if file.Data == nil {
			file.Data = script.New(filename, data)
		} else {
			script := file.Data.(*script.Script)
			script.Code = string(data)
		}
		return nil
	default:
		file.Data = string(data)
		return nil
	}
}

func extractUsername(s string) string {
	slice := strings.Split(s, "\\")
	return slice[len(slice)-1]
}
