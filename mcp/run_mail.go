package mcp

import (
	"fmt"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/smtp"
)

type MailAPI struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Servers     []MailServer `json:"servers"`
	Type        string       `json:"type"`

	info *runtime.MailInfo
}

type MailServer struct {
	Protocol    string `json:"protocol"`
	Host        string `json:"host"`
	Description string `json:"description"`
}

type MailboxSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Mailbox struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Folders     map[string]*MailboxFolder

	info *runtime.MailInfo
}

type MailboxFolder struct {
	Name    string                    `json:"name"`
	Flags   []imap.MailboxFlags       `json:"flags"`
	Folders map[string]*MailboxFolder `json:"folders"`
	Mails   []*smtp.Message           `json:"mails"`

	f *mail.Folder
}

func (m *mokapi) getMailApi(name string) any {
	for _, api := range m.app.Mail.List() {
		cfg := api.Config()
		if cfg != nil && cfg.Info.Name == name {
			result := &MailAPI{
				Name:        name,
				Description: cfg.Info.Description,
				Type:        "mail",
				info:        api,
			}
			for _, server := range cfg.Servers {
				result.Servers = append(result.Servers, MailServer{
					Protocol:    server.Protocol,
					Host:        server.Host,
					Description: server.Description,
				})
			}
			return result
		}
	}
	return nil
}

// notFound is the error for an API whose last file was removed after the object was created.
func (m *MailAPI) notFound() error {
	return fmt.Errorf("mail API '%s' not found", m.Name)
}

func (m *MailAPI) GetMailboxes() ([]MailboxSummary, error) {
	if m.info.Config() == nil {
		return nil, m.notFound()
	}

	result := []MailboxSummary{}
	for _, mb := range m.info.Mailboxes() {
		result = append(result, MailboxSummary{
			Name:        mb.Name,
			Description: mb.Description(),
		})
	}
	return result, nil
}

func (m *MailAPI) GetMailbox(name string) (*Mailbox, error) {
	if m.info.Config() == nil {
		return nil, m.notFound()
	}

	mb, ok := m.info.Mailbox(name)
	if !ok {
		return nil, nil
	}
	username, password := mb.Credentials()
	result := &Mailbox{
		Name:        mb.Name,
		Description: mb.Description(),
		Username:    username,
		Password:    password,
		Folders:     map[string]*MailboxFolder{},
		info:        m.info,
	}
	mb.ReadFolders(func(folders map[string]*mail.Folder) {
		for _, f := range folders {
			result.Folders[f.Name] = getMailboxFolder(f)
		}
	})
	return result, nil
}

func (m *MailAPI) SendMail(to string, msg *smtp.Message) {
	if m.info.Config() == nil {
		panic(m.notFound())
	}
	if mb, ok := m.info.Mailbox(to); !ok {
		panic(fmt.Errorf("mailbox '%s' not found", to))
	} else {
		mb.Append(msg)
	}
}

func getMailboxFolder(f *mail.Folder) *MailboxFolder {
	result := &MailboxFolder{
		Name:    f.Name,
		Flags:   f.Flags,
		Folders: map[string]*MailboxFolder{},
		f:       f,
	}
	for _, sub := range f.Folders {
		result.Folders[sub.Name] = getMailboxFolder(sub)
	}
	for _, msg := range f.Messages {
		result.Mails = append(result.Mails, msg.Message)
	}
	return result
}
