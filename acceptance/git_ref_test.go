package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/dynamic/provider/git/gittest"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	odiePet = `components:
  schemas:
    Pet:
      type: object
      enum:
        - {name: Odie}
`
	garfieldPet = `components:
  schemas:
    Pet:
      type: object
      enum:
        - {name: Garfield}
`
	petPointer     = "#/components/schemas/Pet"
	odieBody       = `{"name":"Odie"}`
	garfieldBody   = `{"name":"Garfield"}`
	petApiTemplate = `openapi: 3.0.0
info:
  title: %s
  version: '1'
servers:
  - url: http://localhost:%d
paths:
  /pet:
    get:
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                $ref: '%s'
`
)

// gitRefBase is what the git $ref suites share: a local Git server, a directory with the OpenAPI
// files a mock developer writes, and the parse errors Mokapi logs for them.
type gitRefBase struct {
	BaseSuite
	git      *gittest.Server
	specsDir string
	tempDir  string
	logs     *errorLog
	hooks    logrus.LevelHooks
}

// prepare starts the Git server and collects the parse errors Mokapi logs. It has to run before
// Mokapi starts, because configured repos are cloned at startup.
func (b *gitRefBase) prepare() {
	b.git = gittest.NewServer(b.T())
	b.specsDir = b.T().TempDir()
	b.tempDir = b.T().TempDir()
	b.logs = &errorLog{}
	b.hooks = logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
	logrus.AddHook(b.logs)
}

// newConfig watches specsDir. The git provider pulls every second and clones into tempDir.
func (b *gitRefBase) newConfig(repos ...static.GitRepo) *static.Config {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: b.specsDir}}
	cfg.Providers.Git = static.GitProvider{PullInterval: "1s", TempDir: b.tempDir, Repositories: repos}
	return cfg
}

func (b *gitRefBase) TearDownSuite() {
	b.BaseSuite.TearDownSuite()
	logrus.StandardLogger().ReplaceHooks(b.hooks)
}

// gitUrl is the $ref URL of a file in a repository of the local Git server, e.g.
// gitUrl("schemas", "file=pet.yaml") is git+https://127.0.0.1:1234/schemas?file=pet.yaml
func (b *gitRefBase) gitUrl(repo, query string) string {
	return "git+" + b.git.Url(repo) + "?" + query
}

// writeApi writes an OpenAPI file, named after the title, whose GET /pet response schema is the
// $ref. It returns the port of the API. Writing the same file again gives the API a new port, which
// changes the file and makes Mokapi parse it again.
func (b *gitRefBase) writeApi(title, ref string) int {
	b.T().Helper()
	return b.writeApiTo(b.specsDir, title, ref)
}

func (b *gitRefBase) writeApiTo(dir, title, ref string) int {
	b.T().Helper()
	port := try.GetFreePort()
	content := fmt.Sprintf(petApiTemplate, title, port, ref)
	b.writeFile(dir, title+".yaml", content)
	return port
}

func (b *gitRefBase) writeFile(dir, name, content string) {
	b.T().Helper()
	b.Require().NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

// waitForPet waits until GET /pet of the API answers with the body.
func (b *gitRefBase) waitForPet(port int, body string) {
	b.T().Helper()
	var last string
	ok := false
	defer func() {
		if !ok {
			b.T().Logf("last response of GET /pet on port %d: %s; errors: %q", port, last, b.logs.about(""))
		}
	}()
	try.WaitFor(b.T(), func() bool {
		res, err := http.Get(fmt.Sprintf("http://localhost:%d/pet", port))
		if err != nil {
			last = err.Error()
			return false
		}
		defer res.Body.Close()
		content, _ := io.ReadAll(res.Body)
		last = fmt.Sprintf("%d %s", res.StatusCode, content)
		return res.StatusCode == http.StatusOK && string(content) == body
	})
	ok = true
}

// waitForRefError waits until Mokapi logs that it could not resolve the $ref to a file of the
// repository and the message contains all the parts. The API and the repository have the same
// name. The API itself is still served, only its schema stays unresolved.
func (b *gitRefBase) waitForRefError(name string, parts ...string) {
	b.T().Helper()
	ok := false
	defer func() {
		if !ok {
			b.T().Logf("errors logged for %s: %q", name, b.logs.about("/"+name+"?"))
		}
	}()
	try.WaitFor(b.T(), func() bool {
		return b.logs.has(append([]string{"resolve reference", "/" + name + "?"}, parts...)...)
	})
	ok = true
	b.False(b.logs.has("unsupported scheme", "/"+name+"?"), "git+https must be a supported scheme")
}

// configUrls lists the URLs of all configs the dashboard shows.
func (b *gitRefBase) configUrls() []string {
	b.T().Helper()
	res, err := http.Get(fmt.Sprintf("http://localhost:%d/api/configs", b.cfg.Api.Port))
	b.Require().NoError(err)
	defer res.Body.Close()
	var configs []struct {
		Url string `json:"url"`
	}
	b.Require().NoError(json.NewDecoder(res.Body).Decode(&configs))
	urls := make([]string, 0, len(configs))
	for _, c := range configs {
		urls = append(urls, c.Url)
	}
	return urls
}

// countConfigs counts the configs of the dashboard whose URL starts with prefix and contains
// all the parts.
func (b *gitRefBase) countConfigs(prefix string, parts ...string) int {
	b.T().Helper()
	count := 0
urls:
	for _, u := range b.configUrls() {
		if !strings.HasPrefix(u, prefix) {
			continue
		}
		for _, p := range parts {
			if !strings.Contains(u, p) {
				continue urls
			}
		}
		count++
	}
	return count
}

// errorLog keeps the error messages Mokapi logs.
type errorLog struct {
	m        sync.Mutex
	messages []string
}

func (l *errorLog) Levels() []logrus.Level {
	return []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
}

func (l *errorLog) Fire(e *logrus.Entry) error {
	l.m.Lock()
	defer l.m.Unlock()
	l.messages = append(l.messages, e.Message)
	return nil
}

// about returns the messages that mention the text.
func (l *errorLog) about(text string) []string {
	l.m.Lock()
	defer l.m.Unlock()
	var found []string
	for _, m := range l.messages {
		if strings.Contains(m, text) {
			found = append(found, m)
		}
	}
	return found
}

func (l *errorLog) has(parts ...string) bool {
	l.m.Lock()
	defer l.m.Unlock()
messages:
	for _, m := range l.messages {
		for _, p := range parts {
			if !strings.Contains(m, p) {
				continue messages
			}
		}
		return true
	}
	return false
}

// GitRefSuite checks what a mock developer can reference with a git+https $ref: files of repos
// that are only cloned for the $ref, and of repos that are configured under providers.git. Every
// scenario has its own API file and its own repository, so the clone counts do not influence
// each other.
type GitRefSuite struct {
	gitRefBase

	httpsRepo      *gittest.Repo
	gitPlusRepo    *gittest.Repo
	pinnedRepo     *gittest.Repo
	otherPinned    *gittest.Repo
	restrictedRepo *gittest.Repo
	// startupPort is the API whose $ref is resolved while Mokapi clones the configured repos
	startupPort int
}

func (suite *GitRefSuite) SetupSuite() {
	suite.prepare()

	// the OpenAPI file is loaded by the provider, so the dashboard lists it with the URL of the repo
	suite.httpsRepo = suite.git.CreateRepo("configured-https", map[string]string{
		"pet.yaml": odiePet,
		"api.yaml": "openapi: 3.0.0\ninfo:\n  title: Configured https\n  version: '1'\n",
	})
	suite.gitPlusRepo = suite.git.CreateRepo("configured-gitplus", map[string]string{"pet.yaml": odiePet})
	suite.pinnedRepo = suite.createRepoWithDevelopBranch("configured-pinned")
	suite.otherPinned = suite.createRepoWithDevelopBranch("configured-other-pinned")
	suite.restrictedRepo = suite.git.CreateRepo("configured-restricted", map[string]string{
		"README.md": "# Schemas",
		"pet.yaml":  odiePet,
	})

	suite.git.CreateRepo("configured-startup", map[string]string{"pet.yaml": odiePet})
	// written before Mokapi starts: Mokapi reads the file while the configured clone is running
	suite.startupPort = suite.writeApi("configured-startup", suite.gitUrl("configured-startup", "file=pet.yaml")+petPointer)

	suite.initCmd(suite.newConfig(
		static.GitRepo{Url: suite.git.Url("configured-startup")},
		static.GitRepo{Url: suite.git.Url("configured-unknown")},
		static.GitRepo{Url: suite.git.Url("configured-https")},
		static.GitRepo{Url: "git+" + suite.git.Url("configured-gitplus")},
		static.GitRepo{Url: "git+" + suite.git.Url("configured-pinned") + "?ref=develop"},
		static.GitRepo{Url: "git+" + suite.git.Url("configured-other-pinned") + "?ref=develop"},
		static.GitRepo{Url: suite.git.Url("configured-restricted"), Files: []string{"README.md"}},
	))
}

// createRepoWithDevelopBranch creates a repo whose default branch has Odie and develop has Garfield.
func (suite *GitRefSuite) createRepoWithDevelopBranch(name string) *gittest.Repo {
	repo := suite.git.CreateRepo(name, map[string]string{"pet.yaml": odiePet})
	repo.CreateBranch("develop")
	repo.Commit(map[string]string{"pet.yaml": garfieldPet})
	repo.Checkout("main")
	return repo
}

func (suite *GitRefSuite) TestRefWithJsonPointerReadsDefaultBranch() {
	suite.git.CreateRepo("default-branch", map[string]string{"pet.yaml": odiePet})

	port := suite.writeApi("default-branch", suite.gitUrl("default-branch", "file=pet.yaml")+petPointer)

	suite.waitForPet(port, odieBody)
}

func (suite *GitRefSuite) TestRefSelectsBranchAndEveryBranchHasItsOwnClone() {
	suite.createRepoWithDevelopBranch("branches")

	mainPort := suite.writeApi("branch-main", suite.gitUrl("branches", "ref=main&file=pet.yaml")+petPointer)
	developPort := suite.writeApi("branch-develop", suite.gitUrl("branches", "ref=develop&file=pet.yaml")+petPointer)

	suite.waitForPet(mainPort, odieBody)
	suite.waitForPet(developPort, garfieldBody)
	suite.Equal(2, suite.git.CloneCount("branches"))
}

func (suite *GitRefSuite) TestFileWorksWithAndWithoutLeadingSlash() {
	suite.git.CreateRepo("slash", map[string]string{"pet.yaml": odiePet})

	withSlash := suite.writeApi("slash-with", suite.gitUrl("slash", "file=/pet.yaml")+petPointer)
	withoutSlash := suite.writeApi("slash-without", suite.gitUrl("slash", "file=pet.yaml")+petPointer)

	suite.waitForPet(withSlash, odieBody)
	suite.waitForPet(withoutSlash, odieBody)
}

func (suite *GitRefSuite) TestSpellingsOfTheSameUrlAreOneEntryInTheDashboard() {
	suite.git.CreateRepo("spelling", map[string]string{"pet.yaml": odiePet})

	a := suite.writeApi("spelling-a", suite.gitUrl("spelling", "ref=main&file=/pet.yaml")+petPointer)
	b := suite.writeApi("spelling-b", suite.gitUrl("spelling", "file=pet.yaml&ref=main")+petPointer)
	c := suite.writeApi("spelling-c", suite.gitUrl("spelling", "file=%2Fpet.yaml&ref=main")+petPointer)

	suite.waitForPet(a, odieBody)
	suite.waitForPet(b, odieBody)
	suite.waitForPet(c, odieBody)
	suite.Equal(1, suite.countConfigs("git+"+suite.git.Url("spelling"), "pet.yaml"))
	suite.Equal(1, suite.git.CloneCount("spelling"))
}

func (suite *GitRefSuite) TestRelativeRefInsideGitFileIsResolvedInTheSameClone() {
	suite.git.CreateRepo("relative", map[string]string{
		"schemas/owner.yaml": "type: object\nrequired: [pet]\nproperties:\n  pet:\n    $ref: './dog.yaml'\n",
		"schemas/dog.yaml":   "type: string\nenum: [Odie]\n",
	})

	port := suite.writeApi("relative", suite.gitUrl("relative", "file=schemas/owner.yaml"))

	suite.waitForPet(port, `{"pet":"Odie"}`)
	suite.Equal(1, suite.git.CloneCount("relative"))
}

func (suite *GitRefSuite) TestRefReusesConfiguredRepoWithHttpsUrl() {
	port := suite.writeApi("reuse-https", suite.gitUrl("configured-https", "file=pet.yaml")+petPointer)

	suite.waitForPet(port, odieBody)
	suite.Equal(1, suite.git.CloneCount("configured-https"))
}

func (suite *GitRefSuite) TestRefReusesConfiguredRepoWithGitPlusHttpsUrl() {
	port := suite.writeApi("reuse-gitplus", suite.gitUrl("configured-gitplus", "file=pet.yaml")+petPointer)

	suite.waitForPet(port, odieBody)
	suite.Equal(1, suite.git.CloneCount("configured-gitplus"))
}

// TestRefResolvedWhileConfiguredRepoIsCloningWaitsForTheClone covers a $ref that is read at startup,
// while the configured clone is still running. Whether the clone is really unfinished at that
// moment depends on timing, the unit tests (TestProvider_ReadWaitsForClone) cover the wait itself.
func (suite *GitRefSuite) TestRefResolvedWhileConfiguredRepoIsCloningWaitsForTheClone() {
	suite.waitForPet(suite.startupPort, odieBody)
	suite.Equal(1, suite.git.CloneCount("configured-startup"))
}

func (suite *GitRefSuite) TestRefToConfiguredRepoThatFailedToCloneReportsCloneError() {
	suite.writeApi("configured-unknown", suite.gitUrl("configured-unknown", "file=pet.yaml")+petPointer)

	suite.waitForRefError("configured-unknown", "repository not found")
}

func (suite *GitRefSuite) TestRefToConfiguredBranchReusesCloneAndIsOneEntryInTheDashboard() {
	port := suite.writeApi("reuse-branch", suite.gitUrl("configured-pinned", "ref=develop&file=/pet.yaml")+petPointer)

	suite.waitForPet(port, garfieldBody)
	suite.Equal(1, suite.git.CloneCount("configured-pinned"))
	// the provider names its file URL file=%2Fpet.yaml&ref=develop, same file as the $ref
	suite.Equal(1, suite.countConfigs("git+"+suite.git.Url("configured-pinned"), "pet.yaml"))
}

func (suite *GitRefSuite) TestRefWithoutRefToConfiguredBranchClonesDefaultBranch() {
	port := suite.writeApi("own-clone", suite.gitUrl("configured-other-pinned", "file=pet.yaml")+petPointer)

	suite.waitForPet(port, odieBody)
	suite.Equal(2, suite.git.CloneCount("configured-other-pinned"), "configured clone of develop and own clone of main")
}

func (suite *GitRefSuite) TestSeveralRefsToTheSameRepoShareOneClone() {
	suite.git.CreateRepo("shared", map[string]string{"pet.yaml": odiePet, "other.yaml": garfieldPet})

	ports := []int{
		suite.writeApi("shared-a", suite.gitUrl("shared", "file=pet.yaml")+petPointer),
		suite.writeApi("shared-b", suite.gitUrl("shared", "file=other.yaml")+petPointer),
		suite.writeApi("shared-c", suite.gitUrl("shared", "file=/pet.yaml")+petPointer),
	}

	suite.waitForPet(ports[0], odieBody)
	suite.waitForPet(ports[1], garfieldBody)
	suite.waitForPet(ports[2], odieBody)
	suite.Equal(1, suite.git.CloneCount("shared"))
}

func (suite *GitRefSuite) TestFilesOfConfiguredRepoDoNotRestrictRef() {
	port := suite.writeApi("restricted", suite.gitUrl("configured-restricted", "file=pet.yaml")+petPointer)

	suite.waitForPet(port, odieBody)
}

func (suite *GitRefSuite) TestHttpsRepoKeepsItsFileUrls() {
	prefix := suite.git.Url("configured-https") + "?"
	try.WaitFor(suite.T(), func() bool {
		return suite.countConfigs(prefix, "file=/api.yaml") == 1
	})
	suite.Equal(0, suite.countConfigs("git+"+suite.git.Url("configured-https")), "provider files are not renamed to git+https")
}

func (suite *GitRefSuite) TestMissingFileParameterIsReported() {
	suite.git.CreateRepo("no-file", map[string]string{"pet.yaml": odiePet})

	suite.writeApi("no-file", suite.gitUrl("no-file", "ref=main")+petPointer)

	suite.waitForRefError("no-file", "git URL requires query parameter 'file'")
}

func (suite *GitRefSuite) TestFileOutsideTheRepositoryIsRejected() {
	suite.git.CreateRepo("escape", map[string]string{"pet.yaml": odiePet})
	// a clone lives in a directory of tempDir, so ../ leads to this file
	suite.writeFile(suite.tempDir, "outside.yaml", odiePet)

	suite.writeApi("escape", suite.gitUrl("escape", "file=../outside.yaml")+petPointer)

	suite.waitForRefError("escape", "outside.yaml")
}

func (suite *GitRefSuite) TestUnknownRepositoryIsReported() {
	suite.writeApi("unknown-repo", suite.gitUrl("unknown-repo", "file=pet.yaml")+petPointer)

	suite.waitForRefError("unknown-repo", "repository not found")
}

func (suite *GitRefSuite) TestUnknownBranchIsReported() {
	suite.git.CreateRepo("unknown-branch", map[string]string{"pet.yaml": odiePet})

	suite.writeApi("unknown-branch", suite.gitUrl("unknown-branch", "ref=nothing&file=pet.yaml")+petPointer)

	suite.waitForRefError("unknown-branch", "refs/heads/nothing")
}

func (suite *GitRefSuite) TestMissingFileInRepositoryIsReported() {
	suite.git.CreateRepo("missing-file", map[string]string{"pet.yaml": odiePet})

	suite.writeApi("missing-file", suite.gitUrl("missing-file", "file=nothing.yaml")+petPointer)

	suite.waitForRefError("missing-file", "nothing.yaml")
}

func (suite *GitRefSuite) TestScriptCanOpenGitFile() {
	suite.git.CreateRepo("testdata", map[string]string{"users.json": `{"name":"Odie"}`})
	port := try.GetFreePort()
	suite.writeFile(suite.specsDir, "open.yaml", fmt.Sprintf(`openapi: 3.0.0
info:
  title: Open
  version: '1'
servers:
  - url: http://localhost:%d
paths:
  /pet:
    get:
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: object
`, port))
	suite.writeFile(suite.specsDir, "open.js", fmt.Sprintf(`import { app } from 'mokapi'

export default function() {
    const user = JSON.parse(open('%s'))
    app.api('Open').http().get('/pet', (req, res) => {
        res.data = user
    })
}
`, suite.gitUrl("testdata", "file=users.json")))

	suite.waitForPet(port, odieBody)
}

// TestShutdownRemovesOnDemandClones stops a Mokapi of its own, so the clones in its temp
// directory can be checked.
func (suite *GitRefSuite) TestShutdownRemovesOnDemandClones() {
	suite.git.CreateRepo("cleanup", map[string]string{"pet.yaml": odiePet})
	specsDir, tempDir := suite.T().TempDir(), suite.T().TempDir()
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: specsDir}}
	cfg.Providers.Git = static.GitProvider{PullInterval: "1s", TempDir: tempDir}
	port := suite.writeApiTo(specsDir, "cleanup", suite.gitUrl("cleanup", "file=pet.yaml")+petPointer)
	cmd, err := Start(cfg)
	suite.Require().NoError(err)

	suite.waitForPet(port, odieBody)
	entries, err := os.ReadDir(tempDir)
	suite.Require().NoError(err)
	suite.NotEmpty(entries, "the on-demand clone lives in tempDir")

	cmd.Stop()

	try.WaitFor(suite.T(), func() bool {
		entries, err := os.ReadDir(tempDir)
		return err == nil && len(entries) == 0
	})
}

// GitRefReloadSuite checks what happens to a referencing OpenAPI file when the referenced file
// changes in the repository, and when a clone failed before. Every scenario has its own
// repository and API file.
type GitRefReloadSuite struct {
	gitRefBase

	liveRepo *gittest.Repo
	livePort int
}

const liveApiTemplate = `openapi: 3.0.0
info:
  title: Configured live
  version: '1'
servers:
  - url: http://localhost:%d
paths:
  /pet:
    get:
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: object
                enum:
                  - {name: %s}
`

func (suite *GitRefReloadSuite) SetupSuite() {
	suite.prepare()
	// an OpenAPI file in a configured https repository, loaded by the provider itself
	suite.livePort = try.GetFreePort()
	suite.liveRepo = suite.git.CreateRepo("configured-live", map[string]string{
		"api.yaml": fmt.Sprintf(liveApiTemplate, suite.livePort, "Odie"),
	})
	suite.initCmd(suite.newConfig(static.GitRepo{Url: suite.git.Url("configured-live")}))
}

// TestChangeOfFileOfConfiguredHttpsRepoKeepsApiServed verifies that a pull, which deletes and
// writes the file again, does not take the API of a configured https repository away.
func (suite *GitRefReloadSuite) TestChangeOfFileOfConfiguredHttpsRepoKeepsApiServed() {
	suite.waitForPet(suite.livePort, odieBody)
	info := suite.cmd.App.Http.Get("Configured live")
	suite.Require().NotNil(info)

	suite.liveRepo.Commit(map[string]string{"api.yaml": fmt.Sprintf(liveApiTemplate, suite.livePort, "Garfield")})

	suite.waitForPet(suite.livePort, garfieldBody)
	suite.Same(info, suite.cmd.App.Http.Get("Configured live"), "the API was not removed and added again")
}

func (suite *GitRefReloadSuite) TestChangeAfterPullShowsNewContent() {
	repo := suite.git.CreateRepo("update", map[string]string{"pet.yaml": odiePet})
	port := suite.writeApi("update", suite.gitUrl("update", "file=pet.yaml")+petPointer)
	suite.waitForPet(port, odieBody)

	repo.Commit(map[string]string{"pet.yaml": garfieldPet})

	suite.waitForPet(port, garfieldBody)
}

func (suite *GitRefReloadSuite) TestDeleteAfterPullFailsReferencingFile() {
	repo := suite.git.CreateRepo("delete", map[string]string{"pet.yaml": odiePet, "README.md": "# Schemas"})
	port := suite.writeApi("delete", suite.gitUrl("delete", "file=pet.yaml")+petPointer)
	suite.waitForPet(port, odieBody)

	repo.Remove("pet.yaml")

	try.WaitFor(suite.T(), func() bool { return suite.logs.has("resolve reference", "/delete?", "pet.yaml") })
}

func (suite *GitRefReloadSuite) TestNextParseTriesAgainAfterFailedClone() {
	suite.writeApi("retry", suite.gitUrl("retry", "file=pet.yaml")+petPointer)
	suite.waitForRefError("retry", "repository not found")

	suite.git.CreateRepo("retry", map[string]string{"pet.yaml": odiePet})
	port := suite.writeApi("retry", suite.gitUrl("retry", "file=pet.yaml")+petPointer)

	suite.waitForPet(port, odieBody)
}
