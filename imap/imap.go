package imap

import (
	"context"
	"strings"
)

const (
	untagged       = "*"
	DateTimeLayout = "02-Jan-2006 15:04:05 -0700"
)

type ConnState uint8

const (
	NotAuthenticatedState ConnState = iota
	AuthenticatedState
	SelectedState
	LogoutState
)

type Handler interface {
	Login(username, password string, ctx context.Context) error
	Select(mailbox string, ctx context.Context) (*Selected, error)
	Unselect(ctx context.Context) error
	List(ref, pattern string, flags []MailboxFlags, ctx context.Context) ([]ListEntry, error)
	Fetch(req *FetchRequest, res FetchResponse, ctx context.Context) error
	Store(req *StoreRequest, res FetchResponse, ctx context.Context) error
}

type Flag string

const (
	// FlagSeen Message has been read
	FlagSeen Flag = "\\Seen"
	// FlagAnswered Message has been answered
	FlagAnswered Flag = "\\Answered"
	// FlagFlagged Message is "flagged" for urgent/special attention
	FlagFlagged Flag = "\\Flagged"
	// FlagDeleted Message is "deleted" for removal by later EXPUNGE
	FlagDeleted Flag = "\\Deleted"
	// FlagDraft Message has not completed composition (marked as a draft).
	FlagDraft Flag = "\\Draft"
	// FlagRecent Message is "recently" arrived in this mailbox.  This session
	// is the first session to have been notified about this
	// message; if the session is read-write, subsequent sessions
	// will not see \Recent set for this message.  This flag can not
	// be altered by the client.
	FlagRecent Flag = "\\Recent"
)

func flagsToString(flags []Flag) string {
	var sb strings.Builder
	for i, f := range flags {
		if i > 0 {
			sb.WriteString(" " + string(f))
		} else {
			sb.WriteString(string(f))
		}
	}
	return sb.String()
}

type MailboxFlags string

const (
	// NoInferiors It is not possible for any child levels of hierarchy to exist
	// under this name; no child levels exist now and none can be
	// created in the future.
	NoInferiors MailboxFlags = "\\Noinferiors"

	// NoSelect It is not possible to use this name as a selectable mailbox.
	NoSelect MailboxFlags = "\\Noselect"

	// Marked The mailbox has been marked "interesting" by the server; the
	// mailbox probably contains messages that have been added since
	// the last time the mailbox was selected.
	Marked MailboxFlags = "\\Marked"

	// UnMarked The mailbox does not contain any additional messages since the
	// last time the mailbox was selected.
	UnMarked MailboxFlags = "\\Unmarked"

	HasNoChildren MailboxFlags = "\\HasNoChildren"

	Subscribed MailboxFlags = "\\Subscribed"

	Trash MailboxFlags = "\\Trash"
)

func joinMailboxFlags(flags []MailboxFlags) string {
	var sb strings.Builder
	for _, f := range flags {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(string(f))
	}
	return sb.String()
}

func joinFlags(flags []Flag) string {
	var sb strings.Builder
	for _, f := range flags {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(string(f))
	}
	return sb.String()
}
