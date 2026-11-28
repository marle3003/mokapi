package mokapi

import (
	"mokapi/engine/common"
	"mokapi/js/mail"
	"mokapi/js/util"

	"github.com/dop251/goja"
)

type Mail struct {
	filter common.MailFilter
	m      *Module
}

type Mailbox struct {
	filter common.MailFilter
	m      *Module
}

func (m *Mail) Mailbox(address string) *Mailbox {
	f := m.filter
	f.Mailbox = address
	return &Mailbox{filter: f, m: m.m}
}

func (m *Mail) Message(do goja.Value, vArgs goja.Value) *Mail {
	onMailMessage(m.m, m.filter, do, vArgs)
	return m
}

func (m *Mailbox) Message(do goja.Value, vArgs goja.Value) *Mailbox {
	onMailMessage(m.m, m.filter, do, vArgs)
	return m
}

func (m *Mail) Send(vMsg goja.Value) *Mail {
	args := prepareMailSend(m.m, m.filter, vMsg)
	if err := m.m.host.MailClient().Send(args); err != nil {
		panic(m.m.vm.ToValue(err.Error()))
	}
	return m
}

// SendAsync resolves the promise with the router, like Send returns it. It is the way to send
// from a handler of a script that is delivering a mail itself.
func (m *Mail) SendAsync(vMsg goja.Value) *goja.Promise {
	args := prepareMailSend(m.m, m.filter, vMsg)
	client := m.m.host.MailClient()
	return util.RunAsync(m.m.vm, m.m.loop,
		func() (struct{}, error) { return struct{}{}, client.Send(args) },
		func(struct{}) (any, error) { return m, nil })
}

func onMailMessage(m *Module, filter common.MailFilter, do goja.Value, vArgs goja.Value) {
	args, err := getOnArgs(m.vm, vArgs)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	f := getHandler(do, args, m.vm, m.loop)
	m.host.OnMail(filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
}

// prepareMailSend converts the message to a plain Go value. It has to run on the event loop
// because it reads from the goja runtime. Invalid input throws.
func prepareMailSend(m *Module, filter common.MailFilter, vMsg goja.Value) *common.MailSendArgs {
	var msg mail.Mail
	if err := m.vm.ExportTo(vMsg, &msg); err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	message, err := msg.ToMessage(m.host)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	return &common.MailSendArgs{Api: filter.Api, Message: message}
}
