package engine

import (
	"mokapi/engine/common"
	"slices"
)

type MailEventHandler = eventHandler[common.MailFilter]

type MailEventDispatcher struct {
	dispatcher[common.MailFilter]
}

func (sh *scriptHost) OnMail(filter common.MailFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.MailEventDispatcher.Register(sh.name, newEventHandler(sh, "smtp", filter, do, args))
}

// EmitSmtp runs the matching handlers on a copy of the mail. Like Kafka, it writes the
// copy back before the caller stores the mail, if a handler did something.
func (e *MailEventDispatcher) EmitSmtp(event *common.MailEvent) []*common.Action {
	message := common.NewMailEventMessage(event.Api, event.Message)
	actions := e.emit(emitOptions[common.MailFilter]{
		shouldEmit: func(f common.MailFilter) bool {
			if !matchesIfPresent(f.Api, event.Api) {
				return false
			}
			return f.Mailbox == "" || slices.Contains(event.Recipients, f.Mailbox)
		},
	}, message)
	if len(actions) > 0 {
		message.ApplyTo(event.Message)
	}
	return actions
}
