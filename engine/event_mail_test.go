package engine_test

import (
	"encoding/json"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/smtp"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func registerMailHandler(d *engine.MailEventDispatcher, name string, filter common.MailFilter, do func(m *common.MailEventMessage)) {
	d.Register(name, &engine.MailEventHandler{
		Filter: filter,
		Execute: func(ctx *common.EventContext) (bool, error) {
			if do != nil {
				do(ctx.Args[0].(*common.MailEventMessage))
			}
			return true, nil
		},
	})
}

func TestMailEventDispatcher_FilterByApi(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	var called []string
	registerMailHandler(d, "all", common.MailFilter{}, func(*common.MailEventMessage) { called = append(called, "all") })
	registerMailHandler(d, "a", common.MailFilter{Api: "Mail A"}, func(*common.MailEventMessage) { called = append(called, "a") })
	registerMailHandler(d, "b", common.MailFilter{Api: "Mail B"}, func(*common.MailEventMessage) { called = append(called, "b") })

	d.EmitSmtp(&common.MailEvent{Api: "Mail A", Message: &smtp.Message{}})

	require.ElementsMatch(t, []string{"all", "a"}, called)
}

func TestMailEventDispatcher_FilterByMailbox_Exact(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	var called []string
	registerMailHandler(d, "alice", common.MailFilter{Mailbox: "alice@foo.bar"}, func(*common.MailEventMessage) { called = append(called, "alice") })
	registerMailHandler(d, "upper", common.MailFilter{Mailbox: "Alice@foo.bar"}, func(*common.MailEventMessage) { called = append(called, "upper") })
	registerMailHandler(d, "bob", common.MailFilter{Mailbox: "bob@foo.bar"}, func(*common.MailEventMessage) { called = append(called, "bob") })

	d.EmitSmtp(&common.MailEvent{Recipients: []string{"alice@foo.bar"}, Message: &smtp.Message{}})

	require.Equal(t, []string{"alice"}, called)
}

func TestMailEventDispatcher_MailboxMatchesBcc(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	var called int
	registerMailHandler(d, "carol", common.MailFilter{Mailbox: "carol@foo.bar"}, func(*common.MailEventMessage) { called++ })

	// the headers don't list carol, only the envelope does
	d.EmitSmtp(&common.MailEvent{
		Recipients: []string{"alice@foo.bar", "carol@foo.bar"},
		Message:    &smtp.Message{To: []smtp.Address{{Address: "alice@foo.bar"}}},
	})

	require.Equal(t, 1, called)
}

func TestMailEventDispatcher_RunsOncePerMailbox(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	calls := map[string]int{}
	for _, name := range []string{"alice@foo.bar", "bob@foo.bar"} {
		registerMailHandler(d, name, common.MailFilter{Mailbox: name}, func(*common.MailEventMessage) { calls[name]++ })
	}

	d.EmitSmtp(&common.MailEvent{
		Recipients: []string{"alice@foo.bar", "bob@foo.bar", "alice@foo.bar"},
		Message:    &smtp.Message{},
	})

	require.Equal(t, map[string]int{"alice@foo.bar": 1, "bob@foo.bar": 1}, calls)
}

func TestMailEventDispatcher_HandlerChangesWrittenBack(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	registerMailHandler(d, "foo", common.MailFilter{}, func(m *common.MailEventMessage) {
		m.Subject = "changed"
		m.Body = "new body"
		m.Encoding = "base64"
		m.Time = time.UnixMilli(5000)
		m.Cc = append(m.Cc, smtp.Address{Address: "cc@foo.bar"})
	})
	msg := &smtp.Message{Subject: "original", Body: "body", Date: time.UnixMilli(1000)}

	actions := d.EmitSmtp(&common.MailEvent{Api: "Mail A", Message: msg})

	require.Len(t, actions, 1)
	require.Equal(t, "changed", msg.Subject)
	require.Equal(t, "new body", msg.Body)
	require.Equal(t, "base64", msg.ContentTransferEncoding)
	require.Equal(t, int64(5000), msg.Date.UnixMilli())
	require.Equal(t, []smtp.Address{{Address: "cc@foo.bar"}}, msg.Cc)
}

func TestMailEventDispatcher_NoHandlerLeavesMailUntouched(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	msg := &smtp.Message{Subject: "original", Server: "smtp://foo", Size: 42}
	registerMailHandler(d, "other", common.MailFilter{Api: "other"}, func(m *common.MailEventMessage) { m.Subject = "changed" })

	actions := d.EmitSmtp(&common.MailEvent{Api: "Mail A", Message: msg})

	require.Empty(t, actions)
	require.Equal(t, &smtp.Message{Subject: "original", Server: "smtp://foo", Size: 42}, msg)
}

func TestMailEventDispatcher_EnvelopeNotChanged(t *testing.T) {
	d := &engine.MailEventDispatcher{}
	registerMailHandler(d, "foo", common.MailFilter{}, func(m *common.MailEventMessage) {
		m.To = []smtp.Address{{Address: "other@foo.bar"}}
		m.Api = "other"
	})
	event := &common.MailEvent{
		Api:        "Mail A",
		Recipients: []string{"alice@foo.bar"},
		Message:    &smtp.Message{Server: "smtp://foo", To: []smtp.Address{{Address: "alice@foo.bar"}}},
	}

	d.EmitSmtp(event)

	require.Equal(t, "Mail A", event.Api)
	require.Equal(t, []string{"alice@foo.bar"}, event.Recipients)
	require.Equal(t, "smtp://foo", event.Message.Server)
	require.Equal(t, []smtp.Address{{Address: "other@foo.bar"}}, event.Message.To, "only the header changes")
}

func TestMailEventDispatcher_FieldNamesTimeEncoding(t *testing.T) {
	m := common.NewMailEventMessage("Mail A", &smtp.Message{
		ContentTransferEncoding: "7bit",
		Date:                    time.UnixMilli(1000),
	})

	t.Run("json", func(t *testing.T) {
		b, err := json.Marshal(m)
		require.NoError(t, err)
		var fields map[string]any
		require.NoError(t, json.Unmarshal(b, &fields))
		require.Equal(t, "Mail A", fields["api"])
		require.Equal(t, "7bit", fields["encoding"])
		require.Contains(t, fields, "time")
		require.NotContains(t, fields, "date")
		require.NotContains(t, fields, "contentTransferEncoding")
	})

	t.Run("javascript", func(t *testing.T) {
		vm := goja.New()
		vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		require.NoError(t, vm.Set("mail", m))
		v, err := vm.RunString(`[mail.api, mail.encoding, mail.time.unixMilli()].join()`)
		require.NoError(t, err)
		require.Equal(t, "Mail A,7bit,1000", v.String())
	})
}
