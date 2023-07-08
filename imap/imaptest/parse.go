package imaptest

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/imap"
	"strconv"
	"strings"
	"time"
)

const (
	dateTimeLayout = "_2-Jan-2006 15:04:05 -0700"

	internalDateKey = "INTERNALDATE"
	sizeKey         = "RFC822.SIZE"
	envelopeKey     = "ENVELOPE"
)

func ParseInternalDate(s string) (time.Time, error) {
	index := strings.Index(s, internalDateKey) + len(internalDateKey)
	r := bufio.NewReader(strings.NewReader(s[index:]))
	if err := expectSP(r); err != nil {
		return time.Time{}, err
	}
	return parseDateTime(r)
}

func ParseSize(s string) (int64, error) {
	index := strings.Index(s, sizeKey) + len(sizeKey)
	r := bufio.NewReader(strings.NewReader(s[index:]))
	if err := expectSP(r); err != nil {
		return 0, err
	}
	return parseInt64(r)
}

func ParseEnvelope(s string) (*imap.Envelope, error) {
	index := strings.Index(s, envelopeKey) + len(envelopeKey)
	r := bufio.NewReader(strings.NewReader(s[index:]))

	if err := expectSP(r); err != nil {
		return nil, err
	}

	if err := expect(r, '('); err != nil {
		return nil, err
	}

	date, err := parseDateTime(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	subject, err := parseString(r)
	if err != nil {
		return nil, errors.Wrap(err, "expected subject")
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	from, err := parseAddressList(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	sender, err := parseAddressList(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	replyTo, err := parseAddressList(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	to, err := parseAddressList(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	cc, err := parseAddressList(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	bcc, err := parseAddressList(r)
	if err != nil {
		return nil, err
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	inReplyTo, err := parseString(r)
	if err != nil {
		return nil, errors.Wrap(err, "expected inReplyTo")
	}

	if err = expectSP(r); err != nil {
		return nil, err
	}

	messageId, err := parseString(r)
	if err != nil {
		return nil, errors.Wrap(err, "expected messageId")
	}

	return &imap.Envelope{
		Date:      date,
		Subject:   subject,
		From:      from,
		Sender:    sender,
		ReplyTo:   replyTo,
		To:        to,
		Cc:        cc,
		Bcc:       bcc,
		InReplyTo: inReplyTo,
		MessageId: messageId,
	}, nil
}

func parseAddressList(r *bufio.Reader) ([]imap.Address, error) {
	if b, err := r.Peek(1); err != nil {
		return nil, err
	} else if b[0] == 'N' {
		return nil, expectNil(r)
	}
	if err := expect(r, '('); err != nil {
		return nil, err
	}
	var list []imap.Address
	for {
		if b, err := r.Peek(1); err != nil {
			return nil, err
		} else if b[0] == ')' {
			_, _ = r.ReadByte()
			break
		}
		addr, err := parseAddress(r)
		if err != nil {
			return nil, err
		}
		list = append(list, addr)
	}
	return list, nil
}

func parseAddress(r *bufio.Reader) (imap.Address, error) {
	if err := expect(r, '('); err != nil {
		return imap.Address{}, err
	}
	name, err := parseString(r)
	if err != nil {
		return imap.Address{}, err
	}
	if err = expectSP(r); err != nil {
		return imap.Address{}, err
	}
	if err = expectNil(r); err != nil {
		return imap.Address{}, err
	}
	if err = expectSP(r); err != nil {
		return imap.Address{}, err
	}
	mailbox, err := parseString(r)
	if err != nil {
		return imap.Address{}, err
	}
	if err = expectSP(r); err != nil {
		return imap.Address{}, err
	}
	host, err := parseString(r)
	if err != nil {
		return imap.Address{}, err
	}

	if err := expect(r, ')'); err != nil {
		return imap.Address{}, err
	}
	return imap.Address{
		Name:    name,
		Mailbox: mailbox,
		Host:    host,
	}, nil
}

func parseString(r *bufio.Reader) (string, error) {
	if b, err := r.Peek(1); err != nil {
		return "", err
	} else if b[0] == 'N' {
		return "", expectNil(r)
	}

	var sb strings.Builder
	if err := expect(r, '"'); err != nil {
		return "", err
	}
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '"' {
			break
		}
		sb.WriteByte(b)
	}
	return sb.String(), nil
}

func parseDateTime(r *bufio.Reader) (time.Time, error) {
	s, err := parseString(r)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(dateTimeLayout, s)
}

func parseInt64(r *bufio.Reader) (int64, error) {
	var sb strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		sb.WriteByte(b)

		if b, err := r.Peek(1); err != nil {
			return 0, err
		} else if b[0] == ' ' {
			break
		}
	}
	return strconv.ParseInt(sb.String(), 10, 64)
}

func expectNil(r *bufio.Reader) error {
	b := [3]byte{}
	n, err := r.Read(b[:])
	if err != nil {
		return err
	}
	if n != 3 || string(b[:]) != "NIL" {
		return fmt.Errorf("expected NIL, got %s", b)
	}
	return nil
}

func expectSP(r *bufio.Reader) error {
	return expect(r, ' ')
}

func expect(r *bufio.Reader, expect byte) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b != expect {
		return fmt.Errorf("expected %q, got %q", expect, b)
	}
	return nil
}
