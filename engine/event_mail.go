package engine

import (
	"mokapi/engine/common"
	"mokapi/smtp"
)

type MailEventHandler = eventHandler[common.MailFilter]

type MailEventDispatcher struct {
	dispatcher[common.MailFilter]
}

func (sh *scriptHost) OnMail(filter common.MailFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.MailEventDispatcher.Register(sh.name, newEventHandler(sh, "smtp", filter, do, args))
}

func (e *MailEventDispatcher) EmitSmtp(message *smtp.Message, status *smtp.Status) []*common.Action {
	return e.emit(emitOptions[common.MailFilter]{}, message, status)
}
