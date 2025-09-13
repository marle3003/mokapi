package runtime

import (
	"fmt"
	"mokapi/providers/mail"
	"mokapi/runtime/search"
	"path"
	"strings"
)

type mailSearchIndexData struct {
	Type          string                  `json:"type"`
	Discriminator string                  `json:"discriminator"`
	Api           string                  `json:"api"`
	Name          string                  `json:"name"`
	Version       string                  `json:"version"`
	Description   string                  `json:"description"`
	Contact       *mail.Contact           `json:"contact"`
	Servers       []mailSearchIndexServer `json:"servers"`
}

type mailSearchIndexServer struct {
	mail.Server
	Name string `json:"name"`
}

type mailSearchIndexMailbox struct {
	Type          string                  `json:"type"`
	Discriminator string                  `json:"discriminator"`
	Api           string                  `json:"api"`
	Name          string                  `json:"name"`
	Username      string                  `json:"username"`
	Password      string                  `json:"password"`
	Description   string                  `json:"description"`
	Folders       []mailSearchIndexFolder `json:"folders"`
}

type mailSearchIndexFolder struct {
	Name  string   `json:"name"`
	Flags []string `json:"flags"`
}

func (s *MailStore) addToIndex(cfg *mail.Config) {
	if cfg == nil || cfg.Info.Name == "" {
		return
	}

	c := mailSearchIndexData{
		Type:          "mail",
		Discriminator: "mail",
		Api:           cfg.Info.Name,
		Name:          cfg.Info.Name,
		Version:       cfg.Info.Version,
		Description:   cfg.Info.Description,
		Contact:       cfg.Info.Contact,
	}
	for name, server := range cfg.Servers {
		if server == nil {
			continue
		}
		c.Servers = append(c.Servers, mailSearchIndexServer{
			Name:   name,
			Server: *server,
		})
	}

	add(s.index, fmt.Sprintf("mail_%s", cfg.Info.Name), c)

	for name, mb := range cfg.Mailboxes {
		mbi := mailSearchIndexMailbox{
			Type:          "mail",
			Discriminator: "mail_mailbox",
			Api:           cfg.Info.Name,
			Name:          name,
			Username:      mb.Username,
			Password:      mb.Password,
			Description:   mb.Description,
		}
		for n, f := range mb.Folders {
			mbi.Folders = append(mbi.Folders, getMailboxFolders(f, n)...)
		}
		add(s.index, fmt.Sprintf("mail_%s_%s", cfg.Info.Name, name), mbi)
	}
}

func getMailSearchResult(fields map[string]string, discriminator []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type: "Mail",
	}

	if len(discriminator) == 1 {
		result.Title = fields["name"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Title,
		}
		return result, nil
	}

	switch discriminator[1] {
	case "mailbox":
		result.Domain = fields["api"]
		result.Title = fields["name"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"mailbox": fields["name"],
		}
	default:
		return result, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}

func (s *MailStore) removeFromIndex(cfg *mail.Config) {
	_ = s.index.Delete(fmt.Sprintf("mail_%s", cfg.Info.Name))

	for name := range cfg.Mailboxes {
		_ = s.index.Delete(fmt.Sprintf("mail_%s_%s", cfg.Info.Name, name))
	}
}

func getMailboxFolders(f *mail.FolderConfig, name string) []mailSearchIndexFolder {
	var result []mailSearchIndexFolder
	result = append(result, mailSearchIndexFolder{
		Name:  name,
		Flags: f.Flags,
	})

	for childName, child := range f.Folders {
		children := getMailboxFolders(child, path.Join(name, childName))
		result = append(result, children...)
	}

	return result
}
