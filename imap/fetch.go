package imap

import (
	"fmt"
	"strconv"
	"strings"
)

type FetchOptions uint

const (
	FetchFlags FetchOptions = 1 << iota
	FetchEnvelope
	FetchInternalDate
	FetchRFC822Header
	FetchRFC822Text
	FetchRFC822Size
	FetchBodyStructure
)

type SequenceSet []Sequence

type Sequence struct {
	start int
	end   int
}

type FetchRequest struct {
	Sequence SequenceSet
	Options  FetchOptions
}

func (c *conn) handleFetch(tag, param string) error {
	args := strings.SplitN(param, " ", 2)
	seq, err := parseFetchSequence(args[0])
	if err != nil {
		return err
	}
	opts, err := parseFetchOptions(args[1])
	if err != nil {
		return err
	}

	req := &FetchRequest{
		Sequence: seq,
		Options:  opts,
	}
	res := fetchResponse{}
	if err = c.handler.Fetch(req, &res, c.ctx); err != nil {
		return err
	}

	if err := c.writeFetchResponse(&res); err != nil {
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
		Sequence{
			start: n,
			end:   n,
		},
	}, nil
}

func parseFetchOptions(s string) (FetchOptions, error) {
	var attr FetchOptions
	switch s {
	case "FAST":
		attr = FetchFlags | FetchInternalDate | FetchRFC822Size
	}
	return attr, nil
}

func (o FetchOptions) Has(opt FetchOptions) bool {
	return o&opt == opt
}

func (c *conn) writeFetchResponse(res *fetchResponse) error {
	for _, msg := range res.messages {
		m := strings.Trim(msg.sb.String(), " ")
		err := c.writeResponse(untagged, &response{
			text: fmt.Sprintf("%v FETCH (%v)", msg.sequenceNumber, m),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
