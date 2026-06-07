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
		if api.Name == name {
			result := &MailAPI{
				Name:        name,
				Description: api.Info.Description,
				Type:        "mail",
				info:        api,
			}
			for _, server := range api.Servers {
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

func (m *MailAPI) GetMailboxes() []MailboxSummary {
	var result []MailboxSummary
	for _, mb := range m.info.Store.Mailboxes {
		result = append(result, MailboxSummary{
			Name:        mb.Name,
			Description: mb.Description,
		})
	}
	return result
}

func (m *MailAPI) GetMailbox(name string) *Mailbox {
	if m.info.Store.Mailboxes == nil {
		return nil
	}
	mb, ok := m.info.Store.Mailboxes[name]
	if !ok {
		return nil
	}
	result := &Mailbox{
		Name:        mb.Name,
		Description: mb.Description,
		Username:    mb.Username,
		Password:    mb.Password,
		Folders:     map[string]*MailboxFolder{},
		info:        m.info,
	}
	for _, f := range mb.Folders {
		result.Folders[f.Name] = getMailboxFolder(f)
	}
	return result
}

func (m *MailAPI) SendMail(to string, msg *smtp.Message) {
	if mb, ok := m.info.Store.Mailboxes[to]; !ok {
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
