package imap

import (
	"context"
	"strings"
)

const (
	untagged       = "*"
	dateTimeLayout = "_2-Jan-2006 15:04:05 -0700"
)

type ConnState uint8

const (
	NotAuthenticatedState ConnState = iota
	AuthenticatedState
	SelectedState
	LogoutState
)

type Handler interface {
	Select(mailbox string, ctx context.Context) (*Selected, error)
	Unselect(ctx context.Context) error
	List(ref, pattern string, ctx context.Context) ([]ListEntry, error)
	Fetch(request *FetchRequest, ctx context.Context) ([]FetchResult, error)
}

type Flag string

const (
	FlagSeen     Flag = "\\Seen"
	FlagAnswered Flag = "\\Answered"
	FlagFlagged  Flag = "\\Flagged"
	FlagDeleted  Flag = "\\Deleted"
	FlagDraft    Flag = "\\Draft"
	FlagRecent   Flag = "\\Recent"
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
