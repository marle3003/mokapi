package npm

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"path/filepath"
	"strings"
)

type Provider struct {
	cfg    static.NpmProvider
	ch     chan *dynamic.Config
	config map[string]static.NpmPackage

	reader file.FSReader
	f      *file.Provider
}

func New(cfg static.NpmProvider) *Provider {
	return &Provider{cfg: cfg, reader: &file.Reader{}, f: file.New(static.FileProvider{})}
}

func NewFS(cfg static.NpmProvider, fs file.FSReader) *Provider {
	return &Provider{cfg: cfg, reader: fs, f: file.NewWithWalker(static.FileProvider{}, fs)}
}

func (p *Provider) Read(u *url.URL) (*dynamic.Config, error) {
	workDir, err := p.reader.GetWorkingDir()
	if err != nil {
		return nil, err
	}

	name := u.Host

	q := u.Query()
	scope := q.Get("scope")
	if len(scope) > 0 {
		name = filepath.Join(scope, name)
	}

	dir, err := p.getPackageDir(name, workDir)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, u.Path)
	fileUrl, err := url.Parse(fmt.Sprintf("file:%v", path))
	if err != nil {
		return nil, err
	}

	c, err := p.f.Read(fileUrl)
	if err != nil {
		return nil, err
	}

	info := dynamic.ConfigInfo{
		Provider: "npm",
		Url:      u,
		Time:     c.Info.Time,
	}

	dynamic.Wrap(info, c)
	return c, nil
}

func (p *Provider) Start(ch chan *dynamic.Config, pool *safe.Pool) error {
	workDir, err := p.reader.GetWorkingDir()
	if err != nil {
		return err
	}

	p.ch = make(chan *dynamic.Config)
	p.config = map[string]static.NpmPackage{}

	pool.Go(func(ctx context.Context) {
		p.forward(ch, ctx)
	})

	for _, pkg := range p.cfg.Packages {
		dir, err := p.getPackageDir(pkg.Name, workDir)
		if err != nil {
			log.Error(err)
			continue
		}
		p.config[dir] = pkg
		p.f.Watch(dir, pool)
	}

	err = p.f.Start(p.ch, pool)
	if err != nil {
		return fmt.Errorf("start file provider failed: %w", err)
	}

	return nil
}

func (p *Provider) forward(ch chan *dynamic.Config, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-p.ch:
			path := c.Info.Url.Path
			if len(c.Info.Url.Opaque) > 0 {
				path = c.Info.Url.Opaque
			}

			for dir, pkg := range p.config {
				if strings.HasPrefix(path, dir) {
					relative := path[len(dir)+1:]
					if skip(relative, pkg) {
						continue
					}

					u, err := toUrl(pkg.Name, relative)
					if err != nil {
						log.Errorf("unable to parse npm url %v: %v", c.Info.Url, err)
						u = c.Info.Url
					}

					info := dynamic.ConfigInfo{
						Provider: "npm",
						Url:      u,
						Time:     c.Info.Time,
					}
					dynamic.Wrap(info, c)

					ch <- c
				}
			}
		}
	}
}

func (p *Provider) getPackageDir(name string, workDir string) (string, error) {
	for len(workDir) > 0 {
		dir := filepath.Join(workDir, "node_modules", name)
		if _, err := p.reader.Stat(dir); err != nil {
			newWorkDir := filepath.Dir(workDir)
			if newWorkDir == workDir {
				break
			}
			workDir = newWorkDir
			continue
		}

		return dir, nil
	}

	for _, folder := range p.cfg.GlobalFolders {
		dir := filepath.Join(folder, name)
		if _, err := p.reader.Stat(dir); err == nil {
			return dir, nil
		}
	}

	return "", fmt.Errorf("module %v not found", name)
}

func skip(path string, pkg static.NpmPackage) bool {
	if len(pkg.Files) == 0 && len(pkg.Include) == 0 {
		return false
	}

	path = filepath.ToSlash(path)

	if contains(pkg.Files, path) {
		return false
	}
	if match(pkg.Include, path) {
		return false
	}

	return true
}

func contains(s []string, v string) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func match(s []string, v string) bool {
	for _, i := range s {
		if file.Match(i, v) {
			return true
		}
	}
	return false
}

func toUrl(pkgName, relative string) (*url.URL, error) {
	query := ""
	if strings.HasPrefix(pkgName, "@") {
		index := strings.Index(pkgName, "/")
		query = fmt.Sprintf("?scope=%v", pkgName[0:index])
	}
	return url.Parse(fmt.Sprintf("npm://%v/%v%v", pkgName, relative, query))
}
