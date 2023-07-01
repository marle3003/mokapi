package imap

import "strings"

type ConnState uint8

const (
	NotAuthenticatedState ConnState = iota
	AuthenticatedState
	SelectedState
	LogoutState
)

type Handler interface {
	Select(mailbox string) (*Selected, error)
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

const (
	untagged = "*"
)
