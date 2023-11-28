package npm

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"path/filepath"
	"strings"
)

type Provider struct {
	cfg    static.NpmProvider
	ch     chan *common.Config
	config map[string]static.NpmPackage

	fs file.FSReader
}

func New(cfg static.NpmProvider) *Provider {
	return &Provider{cfg: cfg, fs: &file.Reader{}}
}

func NewFS(cfg static.NpmProvider, fs file.FSReader) *Provider {
	return &Provider{cfg: cfg, fs: fs}
}

func (p *Provider) Read(_ *url.URL) (*common.Config, error) {
	return nil, fmt.Errorf("not supported")
}

func (p *Provider) Start(ch chan *common.Config, pool *safe.Pool) error {
	workDir, err := p.fs.GetWorkingDir()
	if err != nil {
		return err
	}

	p.ch = make(chan *common.Config)
	p.config = map[string]static.NpmPackage{}

	for _, pkg := range p.cfg.Packages {
		dir, err := p.getPackageDir(pkg.Name, workDir)
		if err != nil {
			log.Error(err)
			continue
		}
		p.config[dir] = pkg
		p.startFileProvider(dir, pool)
	}

	pool.Go(func(ctx context.Context) {
		p.forward(ch, ctx)
	})

	return nil
}

func (p *Provider) forward(ch chan *common.Config, ctx context.Context) {
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
					ch <- c
				}
			}
		}
	}
}

func (p *Provider) getPackageDir(name string, workDir string) (string, error) {
	for len(workDir) > 0 {
		dir := filepath.Join(workDir, "node_modules", name)
		if _, err := p.fs.Stat(dir); err != nil {
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
		if _, err := p.fs.Stat(dir); err == nil {
			return dir, nil
		}
	}

	return "", fmt.Errorf("module %v not found", name)
}

func (p *Provider) startFileProvider(dir string, pool *safe.Pool) {
	f := file.NewWithWalker(static.FileProvider{Directory: dir}, p.fs)
	err := f.Start(p.ch, pool)
	if err != nil {
		log.Errorf("unable to start file provider for git: %v", err)
	}
}

func skip(path string, pkg static.NpmPackage) bool {
	if len(pkg.Files) == 0 && len(pkg.Include) == 0 {
		return false
	}

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
