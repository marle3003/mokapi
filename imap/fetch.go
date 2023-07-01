package imap

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FetchAttribute uint

const (
	FetchFlags FetchAttribute = 1 << iota
	FetchEnvelope
	FetchInternalDate
	FetchRFC822Header
	FetchRFC822Text
	FetchRFC822Size
	FetchBodyStructure
)

type SequenceSet []SequenceNumber

type SequenceNumber struct {
	start int
	end   int
}

type FetchRequest struct {
	Sequence SequenceSet
	Options  FetchAttribute
}

type FetchResult struct {
	SequenceNumber uint32
	UID            uint32
	Flags          []Flag
	InternalDate   time.Time
	Size           int64
}

func (c *conn) handleFetch(tag, param string) error {
	args := strings.SplitN(param, " ", 2)
	seq, err := parseFetchSequence(args[0])
	if err != nil {
		return err
	}
	attr, err := parseFetchAttribute(args[1])
	if err != nil {
		return err
	}
	result, err := c.handler.Fetch(&FetchRequest{
		Sequence: seq,
		Options:  attr,
	}, c.ctx)

	if err := c.writeFetchList(result, attr); err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "",
	})
}

func parseFetchSequence(s string) (SequenceSet, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return SequenceSet{
		SequenceNumber{
			start: n,
			end:   n,
		},
	}, nil
}

func parseFetchAttribute(s string) (FetchAttribute, error) {
	var attr FetchAttribute
	switch s {
	case "FAST":
		attr = FetchFlags | FetchInternalDate | FetchRFC822Size
	}
	return attr, nil
}

func (c *conn) writeFetchList(list []FetchResult, attr FetchAttribute) error {
	for _, result := range list {
		if err := c.writeFetchResult(result, attr); err != nil {
			return err
		}
	}
	return nil
}

func (c *conn) writeFetchResult(result FetchResult, attr FetchAttribute) error {

	return c.writeResponse(untagged, &response{
		text: fmt.Sprintf("%v FETCH (%v)", result.SequenceNumber, result.encode(attr)),
	})
}

func (a FetchAttribute) has(attr FetchAttribute) bool {
	return a&attr == attr
}

func (r FetchResult) encode(attr FetchAttribute) string {
	var sb strings.Builder
	if attr.has(FetchFlags) {
		sb.WriteString("FLAGS ()")
	}
	if attr.has(FetchInternalDate) {
		sb.WriteString(fmt.Sprintf(" INTERNALDATE \"%v\"", r.InternalDate.Format(dateTimeLayout)))
	}
	if attr.has(FetchRFC822Size) {
		sb.WriteString(fmt.Sprintf(" RFC822.SIZE %v", r.Size))
	}
	return sb.String()
}
