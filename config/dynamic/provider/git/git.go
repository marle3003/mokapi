package git

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
)

type repository struct {
	url       string
	repoUrl   string
	localPath string
	ref       string

	repo        *git.Repository
	wt          *git.Worktree
	pullOptions *git.PullOptions
	hash        plumbing.Hash
	// config is the first entry of the configuration for this repository and branch. It gives auth
	// and pull interval.
	config static.GitRepo
	// entries are all configuration entries for this repository and branch (the same URL can be
	// configured several times). A file is loaded if any entry allows it.
	entries []static.GitRepo
	init    bool

	// id identifies the repository, key the repository and its branch, see repositoryKey
	id  string
	key string
	// onDemand is set for clones made for a $ref, they only deal with referenced files
	onDemand bool
	// files is the file provider that watches the clone, nil until the clone exists
	files *file.Provider
	// ready is closed when the clone is done (or has failed, then err is set) and files runs
	ready chan struct{}
	err   error

	refM sync.Mutex
	// refs holds the URLs a file of the clone is referenced by, by local path
	refs map[string]map[string]*url.URL
}

type Provider struct {
	pullInterval string
	tempDir      string
	transport    *transport

	m sync.Mutex
	// repos holds the configured and the on-demand repositories by key, an entry is read and added
	// under m
	repos map[string]*repository
	ch    chan dynamic.ConfigEvent
	pool  *safe.Pool
	ctx   context.Context
}

func New(config static.GitProvider) *Provider {
	repoConfigs := config.Repositories
	for _, u := range config.Urls {
		if len(u) > 0 {
			repoConfigs = append(repoConfigs, static.GitRepo{Url: u})
		}
	}

	if config.PullInterval == "" {
		config.PullInterval = "3m"
	}

	repos := make(map[string]*repository)
	for _, repoConfig := range repoConfigs {
		r, err := newRepository(repoConfig)
		if err != nil {
			log.Errorf("skip git repository %v: %v", repoConfig.Url, err)
			continue
		}
		if existing, ok := repos[r.key]; ok {
			existing.merge(r)
		} else {
			repos[r.key] = r
		}
	}
	for _, r := range repos {
		if r.config.PullInterval == "" {
			r.config.PullInterval = config.PullInterval
		}
	}

	t := newTransport()
	client.InstallProtocol("http", http.NewClient(newClient(t)))
	client.InstallProtocol("https", http.NewClient(newClient(t)))

	return &Provider{
		pullInterval: config.PullInterval,
		tempDir:      config.TempDir,
		transport:    t,
		repos:        repos,
	}
}

// newRepository prepares the clone of a configured repository. The URL may be written with
// the prefix git+ (git+https://...), the clone URL is without it. The file URLs keep the
// URL as configured.
func newRepository(config static.GitRepo) (*repository, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url: %w", err)
	}
	u.Scheme = strings.TrimPrefix(u.Scheme, "git+")

	var ref string
	q := u.Query()
	if len(q) > 0 {
		ref = q.Get("ref")
		q.Del("ref")
		u.RawQuery = q.Encode()
	}

	return &repository{
		url:     config.Url,
		repoUrl: u.String(),
		ref:     ref,
		hash:    plumbing.Hash{},
		config:  config,
		entries: []static.GitRepo{config},
		id:      repositoryId(u),
		key:     repositoryKey(u, ref),
		ready:   make(chan struct{}),
	}, nil
}

// merge adds another configuration entry of the same repository and branch. Auth and pull interval
// of the first entry that sets them win, the file URLs keep the spelling of the first entry.
func (r *repository) merge(other *repository) {
	r.entries = append(r.entries, other.config)
	if r.url != other.url {
		log.Infof("git repository %v is configured as %v too, its files are named after the first", r.url, other.url)
	}
	if r.config.Auth == nil {
		r.config.Auth = other.config.Auth
	} else if other.config.Auth != nil && !reflect.DeepEqual(r.config.Auth, other.config.Auth) {
		log.Warnf("git repository %v is configured with different auth, using the first", other.url)
	}
	if r.config.PullInterval == "" {
		r.config.PullInterval = other.config.PullInterval
	} else if other.config.PullInterval != "" && other.config.PullInterval != r.config.PullInterval {
		log.Warnf("git repository %v is configured with different pull intervals (%v, %v), using the first",
			other.url, r.config.PullInterval, other.config.PullInterval)
	}
}

// repositoryId identifies a repository. The URL is the clone URL, so https://host/org/repo.git
// and git+https://host/org/repo/ are the same repository.
func repositoryId(cloneUrl *url.URL) string {
	return cloneUrl.Host + strings.TrimSuffix(strings.TrimSuffix(cloneUrl.Path, "/"), ".git")
}

// repositoryKey identifies a repository and its branch. An empty ref is the default branch.
func repositoryKey(cloneUrl *url.URL, ref string) string {
	return repositoryId(cloneUrl) + "?ref=" + ref
}

func (p *Provider) Start(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
	p.m.Lock()
	p.ch = ch
	p.pool = pool
	p.ctx = contextOf(pool)
	var configured []*repository
	for _, r := range p.repos {
		configured = append(configured, r)
	}
	p.m.Unlock()

	for _, r := range configured {
		r := r
		pool.Go(func(ctx context.Context) {
			err := p.initRepository(ctx, r, ch, pool)
			if err != nil {
				log.Errorf("init git repository '%v' failed: %v", r.url, err)
			}
		})
	}

	return nil
}

// contextOf returns the context of the pool, that is done when the pool stops.
func contextOf(pool *safe.Pool) context.Context {
	ctxs := make(chan context.Context)
	pool.Go(func(ctx context.Context) {
		ctxs <- ctx
		<-ctx.Done()
	})
	return <-ctxs
}

func (p *Provider) initRepository(ctx context.Context, r *repository, ch chan dynamic.ConfigEvent, pool *safe.Pool) (err error) {
	defer func() {
		r.err = err
		if err != nil {
			p.cleanup(r)
			p.forget(r)
		}
		close(r.ready)
	}()

	r.localPath, err = os.MkdirTemp(p.tempDir, "mokapi_git_*")
	if err != nil {
		return fmt.Errorf("unable to create temp dir: %w", err)
	}

	err = p.transport.addAuth(r)
	if err != nil {
		return err
	}

	err = p.clone(ctx, r)
	if err != nil {
		return err
	}

	chFile := make(chan dynamic.ConfigEvent)
	p.startFileProvider(r, chFile, pool)

	err = startPullInterval(r, pool)
	if err != nil {
		return err
	}
	pool.Go(func(ctx context.Context) {
		p.forward(ctx, r, chFile, ch)
	})
	return nil
}

func (p *Provider) clone(ctx context.Context, r *repository) error {
	options := &git.CloneOptions{
		URL:          r.repoUrl,
		Depth:        1,
		SingleBranch: true,
	}
	if r.ref != "" {
		options.ReferenceName = plumbing.NewBranchReferenceName(r.ref)
	}

	var err error
	r.repo, err = git.PlainCloneContext(ctx, r.localPath, false, options)
	if err != nil {
		return fmt.Errorf("unable to clone git %q: %v", r.repoUrl, err)
	}

	r.wt, err = r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to get git worktree: %v", err.Error())
	}

	r.pullOptions = &git.PullOptions{SingleBranch: true, Depth: 1}
	if r.ref != "" {
		// without the branch, a pull asks for the default branch, which a single-branch clone doesn't have
		r.pullOptions.ReferenceName = plumbing.NewBranchReferenceName(r.ref)
	}
	ref, err := r.repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get git head: %w", err)
	}
	r.hash = ref.Hash()
	return nil
}

// forward sends the events of the clone's file provider to ch. A file of the repository is sent
// as a file of the repository, unless it is not part of the configuration (files, include) or the
// clone is on demand. A referenced file is sent once for every URL it is referenced by, with its
// own copy of the config, because a config can only be wrapped once.
func (p *Provider) forward(ctx context.Context, r *repository, chFile chan dynamic.ConfigEvent, ch chan dynamic.ConfigEvent) {
	for {
		select {
		case <-ctx.Done():
			p.cleanup(r)
			return
		case e := <-chFile:
			for _, evt := range p.eventsOf(r, e) {
				if !safe.Send(ctx, ch, evt) {
					break
				}
			}
		}
	}
}

// eventsOf converts an event of the clone's file provider into the events of the URLs the file
// is known by. The delete event of a referenced file is named by the URL, as the config watcher
// stores configs by URL. The delete event of a file that is only loaded by the provider keeps the
// path as name, which the config watcher ignores: a pull deletes and writes a changed file again,
// and removing the config in between would take its API away (the update replaces the config).
func (p *Provider) eventsOf(r *repository, e dynamic.ConfigEvent) []dynamic.ConfigEvent {
	fileUrl, err := fileUrlOf(e)
	if err != nil {
		log.Errorf("unable to read url of git file %v: %v", e.Name, err)
		return nil
	}

	refs := r.referencesOf(e.Name)
	var urls []*url.URL
	own := !r.onDemand && !skip(e.Name[len(r.localPath)+1:], r)
	ownUrl := getUrl(r, fileUrl)
	if own {
		urls = append(urls, ownUrl)
	} else if e.Event != dynamic.Delete && !r.onDemand {
		log.Debugf("skip file: %v", ownUrl)
	}
	for _, u := range refs {
		if !own || u.String() != ownUrl.String() {
			urls = append(urls, u)
		}
	}

	events := make([]dynamic.ConfigEvent, 0, len(urls))
	for _, u := range urls {
		if e.Event == dynamic.Delete {
			name := e.Name
			if _, referenced := refs[u.String()]; referenced {
				name = u.String()
			}
			events = append(events, dynamic.ConfigEvent{Event: e.Event, Name: name})
			continue
		}
		c := copyOf(e.Config)
		wrapConfig(c, r, u)
		events = append(events, dynamic.ConfigEvent{Event: e.Event, Name: e.Name, Config: c})
	}
	return events
}

func fileUrlOf(e dynamic.ConfigEvent) (*url.URL, error) {
	if e.Config != nil {
		return e.Config.Info.Url, nil
	}
	return file.ParseUrl(e.Name)
}

// copyOf returns a config that can be wrapped independently of c. Refs, listeners and data are
// left out on purpose: the config watcher keeps its own config per URL and only takes the new
// Raw and checksum from the event (see ConfigWatcher.addOrUpdate), then parses it again.
func copyOf(c *dynamic.Config) *dynamic.Config {
	return &dynamic.Config{Info: c.Info, Raw: c.Raw, SourceType: c.SourceType}
}

func (p *Provider) startFileProvider(r *repository, ch chan dynamic.ConfigEvent, pool *safe.Pool) {
	cfg := static.FileProvider{}
	if !r.onDemand {
		// an on-demand clone only watches the referenced files, it does not load the repository
		cfg.Directories = []static.FileConfig{{Path: r.localPath}}
	}
	f := file.New(cfg)
	f.SkipPrefix = append(f.SkipPrefix, ".git")
	r.files = f
	err := f.Start(ch, pool)
	if err != nil {
		log.Errorf("unable to start file provider for git: %v", err)
	}
}

func (p *Provider) cleanup(r *repository) {
	err := os.RemoveAll(r.localPath)
	if err != nil {
		log.Debugf("unable to remove temp dir %q: %v", r.localPath, err.Error())
	}
}

func getUrl(r *repository, file *url.URL) *url.URL {
	path := file.Path
	if len(file.Opaque) > 0 {
		path = file.Opaque
	}
	path = strings.TrimPrefix(path, r.localPath)

	u, _ := url.Parse(r.url)
	path = filepath.ToSlash(path)
	q := u.Query()
	q.Add("file", path)
	u.RawQuery = q.Encode()
	return u
}

// skip reports whether no configuration entry of the repository allows the file. An entry without
// files and include allows all files.
func skip(path string, repo *repository) bool {
	path = filepath.ToSlash(path)
	for _, entry := range repo.entries {
		if len(entry.Files) == 0 && len(entry.Include) == 0 || contains(entry.Files, path) || match(entry.Include, path) {
			return false
		}
	}
	return true
}

func startPullInterval(r *repository, pool *safe.Pool) error {
	if len(r.config.PullInterval) == 0 {
		return nil
	}
	interval, err := time.ParseDuration(r.config.PullInterval)
	if err != nil {
		return fmt.Errorf("unable to parse interval %q: %v", r.config.PullInterval, err)
	}

	ticker := time.NewTicker(interval)

	pool.Go(func(ctx context.Context) {
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pull(r)
			}
		}
	})
	return nil
}

func pull(r *repository) {
	if r.repo == nil {
		return
	}
	err := r.repo.Fetch(&git.FetchOptions{})
	if errors.Is(err, git.ErrForceNeeded) {
		err = r.repo.Fetch(&git.FetchOptions{})
	}
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Errorf("git fetch error %v: %v", r.url, err.Error())
		}
		return
	}

	err = r.wt.Pull(r.pullOptions)
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Errorf("unable to pull: %v", err.Error())
		}
		return
	}

	ref, err := r.repo.Head()
	if err != nil {
		log.Errorf("unable to get head: %v", err.Error())
		return
	}

	hash := ref.Hash()
	if hash != r.hash {
		log.Infof("updated git repository from remote")
		r.hash = hash
	}
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

func wrapConfig(c *dynamic.Config, r *repository, u *url.URL) {
	info := dynamic.ConfigInfo{
		Provider: "git",
		Url:      u,
		// to query git log takes too long
		Time: time.Now(),
	}
	h, err := r.repo.Head()
	if err == nil {
		ref := h.Name()
		info.Tags = append(info.Tags, ref.Short())
	}

	dynamic.Wrap(info, c)
}

func (r *repository) addReference(localPath string, u *url.URL) {
	r.refM.Lock()
	defer r.refM.Unlock()

	if r.refs == nil {
		r.refs = make(map[string]map[string]*url.URL)
	}
	if r.refs[localPath] == nil {
		r.refs[localPath] = make(map[string]*url.URL)
	}
	r.refs[localPath][u.String()] = u
}

// referencesOf returns the URLs a file is referenced by, by URL string.
func (r *repository) referencesOf(localPath string) map[string]*url.URL {
	r.refM.Lock()
	defer r.refM.Unlock()

	return maps.Clone(r.refs[localPath])
}
