package engine

import (
	"errors"
	"fmt"
	"mokapi/engine/common"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/smtp"
	"slices"
	"strings"
)

// MailClient delivers mails from scripts straight into the store of a mail API.
type MailClient struct {
	app     *runtime.App
	emitter common.MailEventEmitter
}

func NewMailClient(app *runtime.App, emitter common.MailEventEmitter) *MailClient {
	return &MailClient{app: app, emitter: emitter}
}

// mailTarget is a mail API with the one config snapshot that is used for the whole delivery.
type mailTarget struct {
	info *runtime.MailInfo
	cfg  *mail.Config
}

func (t mailTarget) title() string {
	return t.cfg.Info.Name
}

func (t mailTarget) hasMailboxFor(recipients []string) bool {
	return slices.ContainsFunc(recipients, func(r string) bool {
		_, ok := t.info.Store.Mailbox(r)
		return ok
	})
}

func (c *MailClient) Send(args *common.MailSendArgs) error {
	from, err := args.Message.EnvelopeSender()
	if err != nil {
		return err
	}
	recipients := envelopeRecipients(args.Message)
	if len(recipients) == 0 {
		// before resolving: without recipients no API can be chosen
		return mail.ErrNoRecipient
	}

	target, err := c.resolve(args.Api, recipients)
	if err != nil {
		return err
	}

	h := mail.NewHandler(target.cfg, target.info.Store, c.emitter, c.app.Events)
	return h.Deliver(from, args.Message, recipients, c.app.Monitor.Mail)
}

func (c *MailClient) resolve(api string, recipients []string) (mailTarget, error) {
	if api != "" {
		return c.scoped(api)
	}
	return c.unscoped(recipients)
}

func (c *MailClient) scoped(api string) (mailTarget, error) {
	// a nil config means the API was removed while this call is running
	if info := c.app.Mail.Get(api); info != nil {
		if cfg := info.Config(); cfg != nil {
			return mailTarget{info: info, cfg: cfg}, nil
		}
	}
	return mailTarget{}, fmt.Errorf("mail API '%v' not found", api)
}

// unscoped chooses the API like the spec says: not with two or more APIs that create
// mailboxes, else the single API that already has a mailbox of a recipient, else the single
// API that creates mailboxes.
func (c *MailClient) unscoped(recipients []string) (mailTarget, error) {
	var all, autoCreating, withMailbox []mailTarget
	for _, info := range c.app.Mail.List() {
		cfg := info.Config()
		if cfg == nil {
			continue
		}
		t := mailTarget{info: info, cfg: cfg}
		all = append(all, t)
		if cfg.Settings.AutoCreatesMailboxes() {
			autoCreating = append(autoCreating, t)
		}
		if t.hasMailboxFor(recipients) {
			withMailbox = append(withMailbox, t)
		}
	}

	switch {
	case len(autoCreating) > 1:
		return mailTarget{}, newAmbiguousError("ambiguous mail API: %v create mailboxes automatically (%v). %v",
			len(autoCreating), titles(autoCreating), useScopedApi)
	case len(withMailbox) > 1:
		return mailTarget{}, newAmbiguousError("ambiguous mail API: recipients have mailboxes in %v. %v",
			titles(withMailbox), useScopedApi)
	case len(withMailbox) == 1:
		return withMailbox[0], nil
	case len(autoCreating) == 1:
		return autoCreating[0], nil
	case len(all) == 0:
		return mailTarget{}, errors.New("no mail API found")
	default:
		return mailTarget{}, fmt.Errorf("no mail API has a mailbox for %v and none creates mailboxes automatically (%v). %v",
			strings.Join(recipients, ", "), titles(all), useScopedApi)
	}
}

const useScopedApi = "Use app.api('...').mail() to choose the mail API"

func titles(targets []mailTarget) string {
	list := make([]string, len(targets))
	for i, t := range targets {
		list[i] = "'" + t.title() + "'"
	}
	slices.Sort(list)
	return strings.Join(list, ", ")
}

// envelopeRecipients are to, cc and bcc without duplicates, so that a mailbox receives one copy.
func envelopeRecipients(m *smtp.Message) []string {
	var recipients []string
	for _, list := range [][]smtp.Address{m.To, m.Cc, m.Bcc} {
		for _, a := range list {
			if !slices.Contains(recipients, a.Address) {
				recipients = append(recipients, a.Address)
			}
		}
	}
	return recipients
}
