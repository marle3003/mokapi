package imap

import (
	"fmt"
	"strconv"
	"strings"
)

type FetchAttributes uint

const (
	FetchFlags FetchAttributes = 1 << iota
	FetchEnvelope
	FetchInternalDate
	FetchRFC822Header
	FetchRFC822Text
	FetchRFC822Size
	FetchBodyStructure
	FetchUID
)

type SequenceSet []Sequence

type Sequence struct {
	Start int
	End   int
}

type FetchBody struct {
	Section      []int
	HeaderFields []string
}

type FetchRequest struct {
	Sequence   SequenceSet
	Attributes FetchAttributes
	// nil means everything
	Body *FetchBody
}

func (c *conn) handleFetch(tag, param string) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	args := strings.SplitN(param, " ", 2)
	seq, err := parseFetchSequence(args[0])
	if err != nil {
		return err
	}
	req, err := parseFetch(args[1])
	if err != nil {
		return err
	}

	req.Sequence = seq
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
	args := strings.Split(s, ":")
	start, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}
	end := start
	if len(args) > 1 {
		if end, err = strconv.Atoi(args[1]); err != nil {
			return nil, err
		}
	}
	return SequenceSet{
		Sequence{
			Start: start,
			End:   end,
		},
	}, nil
}

func (a FetchAttributes) Has(attr FetchAttributes) bool {
	return a&attr == attr
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
