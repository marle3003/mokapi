package git

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
)

type githubTransport struct {
	transToken *ghinstallation.Transport
}

func addGitHubAuth(t *transport, r *repository) error {
	key, err := r.config.Auth.GitHub.PrivateKey.Read("")
	if err != nil {
		return err
	}
	if len(key) == 0 {
		return fmt.Errorf("private key is empty")
	}

	transToken, err := ghinstallation.New(http.DefaultTransport, r.config.Auth.GitHub.AppId, r.config.Auth.GitHub.InstallationId, key)
	if err != nil {
		return err
	}

	t.Add(r.repoUrl, &githubTransport{
		transToken: transToken,
	})

	return nil
}

func (t *githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	token, err := t.transToken.Token(r.Context())
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("x-access-token:%v", token)))
	r.Header.Add("Authorization", fmt.Sprintf("Basic %v", auth))
	return http.DefaultTransport.RoundTrip(r)
}
