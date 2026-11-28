package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/imap"
	"mokapi/smtp/smtptest"
	"mokapi/try"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mailScriptBase is what both mail script suites share: the fixture folder mail-script/ copied to
// a temp dir (the ports are only known at runtime), the control endpoints of script.js and the
// views a mock developer has on the result (dashboard API, IMAP, mail log).
type mailScriptBase struct {
	BaseSuite
	httpPort int
}

// sendRequest is the body of POST /send of script.js.
type sendRequest struct {
	Scope   string         `json:"scope,omitempty"`
	Async   bool           `json:"async,omitempty"`
	Message map[string]any `json:"message"`
}

type sendResult struct {
	Ok             bool   `json:"ok"`
	Error          string `json:"error"`
	ResolvesRouter bool   `json:"resolvesRouter"`
}

// receivedMail is a mail as a handler of script.js has seen it.
type receivedMail struct {
	Handler   string   `json:"handler"`
	Api       string   `json:"api"`
	Subject   string   `json:"subject"`
	MessageId string   `json:"messageId"`
	Time      string   `json:"time"`
	Encoding  string   `json:"encoding"`
	To        []string `json:"to"`
	Cc        []string `json:"cc"`
	Bcc       []string `json:"bcc"`
}

type dashboardMessage struct {
	MessageId string    `json:"messageId"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
	Size      int       `json:"size"`
}

// startFixture copies the given files of mail-script/ into a temp dir, replaces the port
// placeholders (__SMTP_A__, __HTTP__, ...), starts Mokapi on it and waits until the script has
// run and all mail APIs are loaded.
func (s *mailScriptBase) startFixture(ports map[string]int, files ...string) {
	t := s.T()
	dir := t.TempDir()
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join("mail-script", file))
		require.NoError(t, err)
		for name, port := range ports {
			content = bytes.ReplaceAll(content, []byte("__"+name+"__"), []byte(fmt.Sprint(port)))
		}
		require.NoError(t, os.WriteFile(filepath.Join(dir, file), content, 0o600))
	}
	s.httpPort = ports["HTTP"]

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: dir}}
	s.initCmd(cfg)

	waitForScript(t, s.baseUrl())
	for _, file := range files {
		if title, ok := strings.CutPrefix(strings.TrimSuffix(file, ".yml"), "mail-"); ok {
			s.waitForMailApi("Mail " + strings.ToUpper(title))
		}
	}
	for name, port := range ports {
		if strings.HasPrefix(name, "SMTP_") || strings.HasPrefix(name, "IMAP_") {
			try.WaitFor(t, func() bool { return isListening(port) })
		}
	}
}

func (s *mailScriptBase) waitForMailApi(title string) {
	s.T().Helper()
	try.WaitFor(s.T(), func() bool {
		info := s.cmd.App.Mail.Get(title)
		return info != nil && info.Config() != nil
	})
}

func (s *mailScriptBase) baseUrl() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.httpPort)
}

func (s *mailScriptBase) dashboardUrl(path string) string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", s.cfg.Api.Port, path)
}

// send asks script.js to call send() or sendAsync(), like a mock developer's own script would.
func (s *mailScriptBase) send(req sendRequest) sendResult {
	s.T().Helper()
	body, err := json.Marshal(req)
	require.NoError(s.T(), err)
	var result sendResult
	s.postJSON("/send", body, &result)
	return result
}

func (s *mailScriptBase) postJSON(path string, body []byte, v any) {
	s.T().Helper()
	res, err := http.Post(s.baseUrl()+path, "application/json", bytes.NewReader(body))
	require.NoError(s.T(), err)
	defer res.Body.Close()
	require.Equal(s.T(), http.StatusOK, res.StatusCode)
	require.NoError(s.T(), json.NewDecoder(res.Body).Decode(v))
}

func (s *mailScriptBase) getJSON(url string, v any) {
	s.T().Helper()
	res, err := http.Get(url)
	require.NoError(s.T(), err)
	defer res.Body.Close()
	require.Equal(s.T(), http.StatusOK, res.StatusCode, url)
	require.NoError(s.T(), json.NewDecoder(res.Body).Decode(v))
}

// seen returns what the handler of script.js with the given name has seen for a mail subject.
func (s *mailScriptBase) seen(handler, subject string) []receivedMail {
	s.T().Helper()
	var all []receivedMail
	s.getJSON(s.baseUrl()+"/received", &all)
	var result []receivedMail
	for _, m := range all {
		if m.Handler == handler && m.Subject == subject {
			result = append(result, m)
		}
	}
	return result
}

// mailboxes returns the number of mails per mailbox of a mail API, as the dashboard shows it.
func (s *mailScriptBase) mailboxes(api string) map[string]int {
	s.T().Helper()
	var list []struct {
		Name        string `json:"name"`
		NumMessages int    `json:"numMessages"`
	}
	s.getJSON(s.dashboardUrl("/api/services/mail/"+url.PathEscape(api)+"/mailboxes"), &list)
	result := map[string]int{}
	for _, mb := range list {
		result[mb.Name] = mb.NumMessages
	}
	return result
}

func (s *mailScriptBase) messages(api, mailbox string) []dashboardMessage {
	s.T().Helper()
	var list []dashboardMessage
	s.getJSON(s.dashboardUrl("/api/services/mail/"+url.PathEscape(api)+"/mailboxes/"+mailbox+"/messages"), &list)
	return list
}

// storedSubjects returns the subjects in a mailbox of a mail API, as the dashboard shows them.
func (s *mailScriptBase) storedSubjects(api, mailbox string) []string {
	s.T().Helper()
	var subjects []string
	for _, m := range s.messages(api, mailbox) {
		subjects = append(subjects, m.Subject)
	}
	return subjects
}

// logEvents returns the data of the mail log entries of a mail API.
func (s *mailScriptBase) logEvents(api string) []map[string]any {
	s.T().Helper()
	var page struct {
		Items []struct {
			Data map[string]any `json:"data"`
		} `json:"items"`
	}
	s.getJSON(s.dashboardUrl("/api/events?traits[namespace]=mail&traits[name]="+url.QueryEscape(api)+"&size=1000"), &page)
	var result []map[string]any
	for _, e := range page.Items {
		result = append(result, e.Data)
	}
	return result
}

func (s *mailScriptBase) logEventFor(api, subject string) map[string]any {
	s.T().Helper()
	for _, e := range s.logEvents(api) {
		if e["subject"] == subject {
			return e
		}
	}
	s.T().Fatalf("no mail log entry with subject %q in %q", subject, api)
	return nil
}

// imapHeaders returns the Subject headers of all mails in the INBOX of an IMAP user.
func (s *mailScriptBase) imapHeaders(port int, user, password string) []string {
	t := s.T()
	t.Helper()
	c := imap.NewClient(fmt.Sprintf("127.0.0.1:%d", port))
	defer c.Close()
	_, err := c.Dial()
	require.NoError(t, err)
	require.NoError(t, c.PlainAuth("", user, password))
	sel, err := c.Select("INBOX", true)
	require.NoError(t, err)
	if sel.NumMessages == 0 {
		return nil
	}
	cmd, err := c.FetchRaw("FETCH 1:* (BODY.PEEK[HEADER.FIELDS (subject)])")
	require.NoError(t, err)
	var headers []string
	for _, m := range cmd.Messages {
		for _, b := range m.Body {
			headers = append(headers, b.Data)
		}
	}
	return headers
}

func message(from, to, subject string) map[string]any {
	return map[string]any{
		"from":    []map[string]string{{"address": from}},
		"to":      []map[string]string{{"address": to}},
		"subject": subject,
		"body":    "body of " + subject,
	}
}

func addressList(addresses ...string) []map[string]string {
	var list []map[string]string
	for _, a := range addresses {
		list = append(list, map[string]string{"address": a})
	}
	return list
}

func smtpUrl(port int) string {
	return fmt.Sprintf("smtp://127.0.0.1:%d", port)
}

// MailScriptSuite is a mock developer who scripts three mail APIs with app.mail():
//   - Mail A creates mailboxes on delivery and has a recipient rule
//   - Mail B has configured mailboxes, a subject rule, maxRecipients 2 and rejects unknown senders
//   - Mail C has an IMAP server only, no SMTP server
type MailScriptSuite struct {
	mailScriptBase
	smtpA, smtpB, imapB, imapC int
}

func (suite *MailScriptSuite) SetupSuite() {
	suite.smtpA, suite.smtpB = try.GetFreePort(), try.GetFreePort()
	suite.imapB, suite.imapC = try.GetFreePort(), try.GetFreePort()
	suite.startFixture(map[string]int{
		"HTTP":   try.GetFreePort(),
		"SMTP_A": suite.smtpA,
		"SMTP_B": suite.smtpB,
		"IMAP_B": suite.imapB,
		"IMAP_C": suite.imapC,
	}, "api.yml", "script.js", "mail-a.yml", "mail-b.yml", "mail-c.yml")
}

// TestHandlerReceivesSmtpMailWithApi verifies that app.mail().message() runs for mails received
// over SMTP in every mail API and that the mail names the API.
func (suite *MailScriptSuite) TestHandlerReceivesSmtpMailWithApi() {
	t := suite.T()
	require.NoError(t, smtptest.SendMail("sender@x.io", "alice@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("smtp to a")))
	require.NoError(t, smtptest.SendMail("bob@b.io", "alice@b.io", smtpUrl(suite.smtpB), smtptest.WithSubject("smtp to b")))

	mailA := suite.seen("all", "smtp to a")
	require.Len(t, mailA, 1)
	require.Equal(t, "Mail A", mailA[0].Api)
	require.Equal(t, []string{"alice@a.io"}, mailA[0].To)

	mailB := suite.seen("all", "smtp to b")
	require.Len(t, mailB, 1)
	require.Equal(t, "Mail B", mailB[0].Api)
}

// TestHandlerScopedByApi verifies that app.api(title).mail().message() runs only for mails of that
// API, also when the handler is limited to a mailbox of it.
func (suite *MailScriptSuite) TestHandlerScopedByApi() {
	t := suite.T()
	require.NoError(t, smtptest.SendMail("sender@x.io", "alice@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("scoped a")))
	require.NoError(t, smtptest.SendMail("bob@b.io", "alice@b.io", smtpUrl(suite.smtpB), smtptest.WithSubject("scoped b")))

	require.Empty(t, suite.seen("api-b", "scoped a"))
	forB := suite.seen("api-b", "scoped b")
	require.Len(t, forB, 1)
	require.Equal(t, "Mail B", forB[0].Api)

	require.Len(t, suite.seen("b-alice", "scoped b"), 1)
	require.Empty(t, suite.seen("b-alice", "scoped a"))
}

// TestMailboxHandlerRunsOnceForToCcBcc verifies that mailbox() matches the envelope recipients
// to, cc and bcc and that the handler of each mailbox runs once per mail.
func (suite *MailScriptSuite) TestMailboxHandlerRunsOnceForToCcBcc() {
	t := suite.T()
	msg := message("sender@x.io", "alice@a.io", "to cc bcc")
	msg["cc"] = addressList("bob@a.io")
	msg["bcc"] = addressList("carol@a.io")

	result := suite.send(sendRequest{Scope: "Mail A", Message: msg})
	require.True(t, result.Ok, result.Error)

	for _, handler := range []string{"all", "alice", "bob", "carol"} {
		require.Len(t, suite.seen(handler, "to cc bcc"), 1, "handler %s", handler)
	}
}

// TestMailboxMatchIsExact verifies that mailbox() compares the address exactly: a handler for
// Alice@A.io does not run for a mail to alice@a.io.
func (suite *MailScriptSuite) TestMailboxMatchIsExact() {
	t := suite.T()
	result := suite.send(sendRequest{Scope: "Mail A", Message: message("sender@x.io", "alice@a.io", "exact match")})
	require.True(t, result.Ok, result.Error)

	require.Len(t, suite.seen("alice", "exact match"), 1)
	require.Empty(t, suite.seen("alice-upper", "exact match"))
}

// TestHandlerChangesStoredMail verifies that a subject changed by a handler is what the dashboard
// and an IMAP client see, while a changed recipient list does not change who gets the mail.
func (suite *MailScriptSuite) TestHandlerChangesStoredMail() {
	t := suite.T()
	require.NoError(t, smtptest.SendMail("sender@x.io", "alice@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("rewrite: in a")))
	require.NoError(t, smtptest.SendMail("bob@b.io", "alice@b.io", smtpUrl(suite.smtpB), smtptest.WithSubject("rewrite: in b")))

	require.Contains(t, suite.storedSubjects("Mail A", "alice@a.io"), "rewritten rewrite: in a")
	require.NotContains(t, suite.storedSubjects("Mail A", "alice@a.io"), "rewrite: in a")
	require.NotContains(t, suite.mailboxes("Mail A"), "other@a.io")

	require.Contains(t, strings.Join(suite.imapHeaders(suite.imapB, "alice", "secret"), "\n"), "rewritten rewrite: in b")

	// the mail log shows the mail as stored, not as received
	suite.logEventFor("Mail A", "rewritten rewrite: in a")
	for _, e := range suite.logEvents("Mail A") {
		require.NotEqual(t, "rewrite: in a", e["subject"])
	}
}

// TestOnSmtpGetsApiAndTimeEncoding verifies that on('smtp') handlers get the same mail as the
// router: the API title, the time as Date and the encoding.
func (suite *MailScriptSuite) TestOnSmtpGetsApiAndTimeEncoding() {
	t := suite.T()
	msg := message("sender@x.io", "alice@a.io", "on smtp fields")
	msg["encoding"] = "7bit"
	result := suite.send(sendRequest{Scope: "Mail A", Message: msg})
	require.True(t, result.Ok, result.Error)
	require.NoError(t, smtptest.SendMail("sender@x.io", "alice@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("on smtp over smtp")))

	sent := suite.seen("on-smtp", "on smtp fields")
	require.Len(t, sent, 1)
	require.Equal(t, "Mail A", sent[0].Api)
	require.Equal(t, "7bit", sent[0].Encoding)
	suite.requireRecent(sent[0].Time)

	viaSmtp := suite.seen("on-smtp", "on smtp over smtp")
	require.Len(t, viaSmtp, 1)
	require.Equal(t, "Mail A", viaSmtp[0].Api)
	suite.requireRecent(viaSmtp[0].Time)
}

func (suite *MailScriptSuite) requireRecent(s string) {
	suite.T().Helper()
	require.NotEmpty(suite.T(), s, "time is not a Date")
	parsed, err := time.Parse(time.RFC3339, s)
	require.NoError(suite.T(), err)
	require.WithinDuration(suite.T(), time.Now(), parsed, time.Minute)
}

// TestSendStoresMailWithoutSmtpServer verifies that send() delivers into the store of a mail API
// that has no SMTP server, and that an IMAP client, the dashboard and the mail log see the mail.
func (suite *MailScriptSuite) TestSendStoresMailWithoutSmtpServer() {
	t := suite.T()
	result := suite.send(sendRequest{Scope: "Mail C", Message: message("sender@x.io", "carol@c.io", "sent into c")})
	require.True(t, result.Ok, result.Error)

	require.Equal(t, []string{"sent into c"}, suite.storedSubjects("Mail C", "carol@c.io"))
	require.Contains(t, strings.Join(suite.imapHeaders(suite.imapC, "carol", "secret"), "\n"), "sent into c")

	event := suite.logEventFor("Mail C", "sent into c")
	require.Equal(t, "", event["error"])
	require.Equal(t, []any{"carol@c.io"}, event["to"])

	var detail struct{ Data dashboardMessage }
	suite.getJSON(suite.dashboardUrl("/api/services/mail/messages/"+url.PathEscape(suite.messages("Mail C", "carol@c.io")[0].MessageId)), &detail)
	require.Greater(t, detail.Data.Size, 0, "size shown in the dashboard")
}

// TestSendRunsHandlersAndOnSmtp verifies that a mail from send() runs the router handlers and the
// on('smtp') handlers, like a mail received over SMTP.
func (suite *MailScriptSuite) TestSendRunsHandlersAndOnSmtp() {
	t := suite.T()
	result := suite.send(sendRequest{Scope: "Mail A", Message: message("sender@x.io", "alice@a.io", "sent runs handlers")})
	require.True(t, result.Ok, result.Error)

	for _, handler := range []string{"all", "alice", "on-smtp"} {
		handled := suite.seen(handler, "sent runs handlers")
		require.Len(t, handled, 1, "handler %s", handler)
		require.Equal(t, "Mail A", handled[0].Api)
	}
	require.Empty(t, suite.seen("api-b", "sent runs handlers"))
}

// TestSendGeneratesMessageIdAndTime verifies that a message without messageId and time gets both,
// in the handler and in the stored mail.
func (suite *MailScriptSuite) TestSendGeneratesMessageIdAndTime() {
	t := suite.T()
	result := suite.send(sendRequest{Scope: "Mail A", Message: message("sender@x.io", "generated@a.io", "generated fields")})
	require.True(t, result.Ok, result.Error)

	handled := suite.seen("all", "generated fields")
	require.Len(t, handled, 1)
	require.NotEmpty(t, handled[0].MessageId)
	suite.requireRecent(handled[0].Time)

	stored := suite.messages("Mail A", "generated@a.io")
	require.Len(t, stored, 1)
	require.Equal(t, handled[0].MessageId, stored[0].MessageId)
	require.WithinDuration(t, time.Now(), stored[0].Date, time.Minute)
}

// TestSendRejected verifies that send() throws what a real SMTP client would be told and that a
// rejected mail leaves nothing behind: no mail in a mailbox and no new mailbox, also not for the
// recipients that would have been fine.
func (suite *MailScriptSuite) TestSendRejected() {
	testcases := []struct {
		name    string
		api     string
		message map[string]any
		error   string
	}{
		{
			name:    "rule with reject message",
			api:     "Mail B",
			message: message("alice@b.io", "bob@b.io", "forbidden offer"),
			error:   "Subject not allowed by Mail B",
		},
		{
			name:    "unknown sender",
			api:     "Mail B",
			message: message("stranger@x.io", "bob@b.io", "unknown sender"),
			error:   "Unknown mailbox stranger@x.io",
		},
		{
			name: "too many recipients",
			api:  "Mail B",
			message: func() map[string]any {
				m := message("alice@b.io", "alice@b.io", "too many")
				m["cc"] = addressList("bob@b.io")
				m["bcc"] = addressList("dave@b.io")
				return m
			}(),
			error: "Too many recipients of 2 reached",
		},
		{
			name:    "unknown recipient without autoCreateMailbox",
			api:     "Mail B",
			message: message("alice@b.io", "ghost@b.io", "unknown recipient"),
			error:   "Unknown mailbox ghost@b.io",
		},
		{
			name: "recipient rule rejects the second recipient",
			api:  "Mail A",
			message: func() map[string]any {
				m := message("sender@x.io", "fine@a.io", "one recipient blocked")
				m["cc"] = addressList("blocked@a.io")
				return m
			}(),
			error: "Recipient blocked by Mail A",
		},
	}

	for _, tc := range testcases {
		suite.T().Run(tc.name, func(t *testing.T) {
			before := suite.mailboxes(tc.api)

			result := suite.send(sendRequest{Scope: tc.api, Message: tc.message})

			require.False(t, result.Ok)
			require.Contains(t, result.Error, tc.error)
			require.Equal(t, before, suite.mailboxes(tc.api))
		})
	}
}

// TestSendAsyncResolvesAndRejects verifies that sendAsync() resolves with the router when the mail
// is delivered and rejects with the same error as send().
func (suite *MailScriptSuite) TestSendAsyncResolvesAndRejects() {
	t := suite.T()

	ok := suite.send(sendRequest{Scope: "Mail A", Async: true, Message: message("sender@x.io", "async@a.io", "async delivered")})
	require.True(t, ok.Ok, ok.Error)
	require.True(t, ok.ResolvesRouter)
	require.Equal(t, []string{"async delivered"}, suite.storedSubjects("Mail A", "async@a.io"))

	rejected := suite.send(sendRequest{Scope: "Mail B", Async: true, Message: message("alice@b.io", "bob@b.io", "async forbidden")})
	require.False(t, rejected.Ok)
	require.Contains(t, rejected.Error, "Subject not allowed by Mail B")
}

// TestSendUnknownApiTitle verifies that send() names the title when no mail API has it.
func (suite *MailScriptSuite) TestSendUnknownApiTitle() {
	result := suite.send(sendRequest{Scope: "Mail Missing", Message: message("sender@x.io", "alice@a.io", "to nowhere")})

	require.False(suite.T(), result.Ok)
	require.Contains(suite.T(), result.Error, "mail API 'Mail Missing' not found")
}

// TestUnscopedSendResolution verifies how app.mail().send() finds the mail API: an existing mailbox
// wins over autoCreateMailbox, the API with autoCreateMailbox takes mails for unknown recipients
// and recipients in two APIs are an error that names both and points to app.api().
func (suite *MailScriptSuite) TestUnscopedSendResolution() {
	suite.T().Run("existing mailbox wins", func(t *testing.T) {
		before := suite.mailboxes("Mail A")

		result := suite.send(sendRequest{Message: message("alice@b.io", "bob@b.io", "unscoped existing")})

		require.True(t, result.Ok, result.Error)
		require.Contains(t, suite.storedSubjects("Mail B", "bob@b.io"), "unscoped existing")
		require.Equal(t, before, suite.mailboxes("Mail A"))
	})

	suite.T().Run("falls back to the API with autoCreateMailbox", func(t *testing.T) {
		result := suite.send(sendRequest{Message: message("sender@x.io", "fresh@a.io", "unscoped fallback")})

		require.True(t, result.Ok, result.Error)
		require.Equal(t, []string{"unscoped fallback"}, suite.storedSubjects("Mail A", "fresh@a.io"))
	})

	suite.T().Run("recipients in two APIs", func(t *testing.T) {
		msg := message("sender@x.io", "bob@b.io", "unscoped ambiguous")
		msg["cc"] = addressList("carol@c.io")
		bobBefore := suite.storedSubjects("Mail B", "bob@b.io")
		carolBefore := suite.storedSubjects("Mail C", "carol@c.io")

		result := suite.send(sendRequest{Message: msg})

		require.False(t, result.Ok)
		require.Contains(t, result.Error, "Mail B")
		require.Contains(t, result.Error, "Mail C")
		require.Contains(t, result.Error, "app.api(")
		require.Equal(t, bobBefore, suite.storedSubjects("Mail B", "bob@b.io"))
		require.Equal(t, carolBefore, suite.storedSubjects("Mail C", "carol@c.io"))
	})
}

// TestSmtpRejectedRecipientCreatesNoMailbox verifies that a recipient rejected by a rule leaves no
// mailbox behind: Mokapi creates mailboxes when the mail is delivered, not at RCPT.
func (suite *MailScriptSuite) TestSmtpRejectedRecipientCreatesNoMailbox() {
	t := suite.T()
	err := smtptest.SendMail("sender@x.io", "blocked@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("to blocked"))

	require.ErrorContains(t, err, "Recipient blocked by Mail A")
	require.NotContains(t, suite.mailboxes("Mail A"), "blocked@a.io")
}

// TestHandlerActionsAppearInMailLog verifies that tags, priority and track show up in the actions
// of the mail log entry: only handlers that track the mail appear, highest priority first.
func (suite *MailScriptSuite) TestHandlerActionsAppearInMailLog() {
	t := suite.T()
	require.NoError(t, smtptest.SendMail("sender@x.io", "alice@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("track: logged")))
	require.NoError(t, smtptest.SendMail("sender@x.io", "alice@a.io", smtpUrl(suite.smtpA), smtptest.WithSubject("not tracked")))

	tracked := suite.logEventFor("Mail A", "track: logged")
	var order []any
	for _, a := range tracked["actions"].([]any) {
		order = append(order, a.(map[string]any)["tags"].(map[string]any)["order"])
	}
	require.Equal(t, []any{"first", "second"}, order)

	require.Empty(t, suite.logEventFor("Mail A", "not tracked")["actions"])
}

// MailScriptTwoAutoCreateSuite covers the unscoped send() of a mock developer who has two mail APIs
// with autoCreateMailbox: Mokapi can't know where a new mailbox belongs.
type MailScriptTwoAutoCreateSuite struct {
	mailScriptBase
}

func (suite *MailScriptTwoAutoCreateSuite) SetupSuite() {
	suite.startFixture(map[string]int{
		"HTTP":   try.GetFreePort(),
		"SMTP_A": try.GetFreePort(),
		"SMTP_D": try.GetFreePort(),
	}, "api.yml", "script.js", "mail-a.yml", "mail-d.yml")
}

// TestUnscopedSendWithTwoAutoCreateApisFails verifies that send() without app.api() throws with two
// mail APIs that create mailboxes, names both and delivers nothing.
func (suite *MailScriptTwoAutoCreateSuite) TestUnscopedSendWithTwoAutoCreateApisFails() {
	t := suite.T()
	result := suite.send(sendRequest{Message: message("sender@x.io", "new@a.io", "which api")})

	require.False(t, result.Ok)
	require.Contains(t, result.Error, "Mail A")
	require.Contains(t, result.Error, "Mail D")
	require.Contains(t, result.Error, "app.api(")
	require.NotContains(t, suite.mailboxes("Mail A"), "new@a.io")
	require.NotContains(t, suite.mailboxes("Mail D"), "new@a.io")
}
