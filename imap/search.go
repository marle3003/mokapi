package imap

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"time"
)

const SearchDateLayout = "2-Jan-2006"

type SearchRequest struct {
	Criteria *SearchCriteria
}

type SearchCriteria struct {
	Seq Set
	UID Set

	Flag    []string
	NotFlag []string

	Before     time.Time
	Since      time.Time
	SentBefore time.Time
	SentSince  time.Time

	Body []string
	Text []string

	Larger  int64
	Smaller int64

	Not []SearchCriteria
	Or  [][2]SearchCriteria

	Headers []HeaderCriteria
}

type HeaderCriteria struct {
	Name  string
	Value string
}

type SearchResponse struct {
	All Set
}

func (c *conn) handleSearch(tag string, d *Decoder, isUid bool) error {
	r := &SearchRequest{
		Criteria: &SearchCriteria{},
	}

	for {
		err := readSearchKey(r.Criteria, d)
		if err != nil {
			return err
		}
		if !d.IsSP() {
			break
		}
		d.SP()
	}

	res, err := c.handler.Search(r, c.ctx)
	if err != nil {
		return c.writeResponse(tag, &response{
			status: no,
			text:   err.Error(),
		})
	}
	err = c.writeSearchResponse(res)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "SEARCH completed",
	})
}

func readSearchKey(criteria *SearchCriteria, d *Decoder) error {
	if d.IsList() {
		return d.List(func() error {
			return readSearchKey(criteria, d)
		})
	}
	return readAtomSearchKey(criteria, d)
}

func readAtomSearchKey(criteria *SearchCriteria, d *Decoder) error {
	key := strings.ToUpper(d.Read(func(r byte) bool {
		return r == '*' || isAtom(r)
	}))
	if key == "" {
		return nil
	}
	switch key {
	case "ALL":
		// do nothing
	case "ANSWERED", "DELETED", "DRAFT", "FLAGGED", "RECENT", "SEEN":
		criteria.Flag = append(criteria.Flag, key)
	case "UNANSWERED", "UNDELETED", "UNDRAFT", "UNFLAGGED", "UNSEEN":
		criteria.NotFlag = append(criteria.NotFlag, strings.TrimPrefix(key, "UN"))
	case "NEW":
		criteria.Flag = append(criteria.Flag, "RECENT")
		criteria.NotFlag = append(criteria.NotFlag, "SEEN")
	case "OLD":
		criteria.NotFlag = append(criteria.NotFlag, "RECENT")
	case "KEYWORD", "UNKEYWORD":
		flag, err := d.SP().ReadFlag()
		if err != nil {
			return err
		}
		if key == "KEYWORD" {
			criteria.Flag = append(criteria.Flag, flag)
		} else {
			criteria.NotFlag = append(criteria.Flag, flag)
		}
	case "BCC", "CC", "FROM", "TO", "SUBJECT":
		value, err := d.SP().String()
		if err != nil {
			return err
		}
		hc := HeaderCriteria{
			Name:  cases.Title(language.English).String(strings.ToLower(key)),
			Value: value,
		}
		criteria.Headers = append(criteria.Headers, hc)
	case "BEFORE", "ON", "SENTBEFORE", "SENTON", "SENTSINCE", "SINCE":
		value, err := d.SP().String()
		if err != nil {
			return err
		}
		date, err := time.Parse(SearchDateLayout, value)
		if err != nil {
			return err
		}

		op := &SearchCriteria{}
		switch key {
		case "BEFORE":
			op.Before = date
		case "ON":
			op.Since = date
			op.Before = date.Add(24 * time.Hour)
		case "SENTBEFORE":
			op.SentSince = date
		case "SENTON":
			op.SentSince = date
			op.SentBefore = date.Add(24 * time.Hour)
		case "SENTSINCE":
			op.SentSince = date
		case "SINCE":
			op.Since = date
		}
		criteria.And(op)
	case "BODY":
		value, err := d.SP().String()
		if err != nil {
			return err
		}
		criteria.Body = append(criteria.Body, value)
	case "HEADER":
		hc := HeaderCriteria{}
		var err error
		hc.Name, err = d.SP().String()
		if err != nil {
			return err
		}
		hc.Value, err = d.SP().String()
		if err != nil {
			return err
		}
		criteria.Headers = append(criteria.Headers, hc)
	case "LARGER", "SMALLER":
		v, err := d.SP().Int64()
		if err != nil {
			return err
		}
		if key == "LARGER" {
			criteria.And(&SearchCriteria{Larger: v})
		} else {
			criteria.And(&SearchCriteria{Smaller: v})
		}
	case "NOT":
		not := SearchCriteria{}
		err := readSearchKey(&not, d.SP())
		if err != nil {
			return err
		}
		criteria.Not = append(criteria.Not, not)
	case "OR":
		var or [2]SearchCriteria
		if err := readSearchKey(&or[0], d.SP()); err != nil {
			return err
		}
		if err := readSearchKey(&or[1], d.SP()); err != nil {
			return err
		}
		criteria.Or = append(criteria.Or, or)
	case "TEXT":
		value, err := d.SP().String()
		if err != nil {
			return err
		}
		criteria.Text = append(criteria.Text, value)
	default:
		set, err := parseSequence(key)
		if err != nil {
			return err
		}
		criteria.Seq = &set
	}
	return nil
}

func (c *conn) writeSearchResponse(res *SearchResponse) error {
	if res.All == nil {
		return nil
	}
	s := res.All.String()
	if s != "" {
		err := c.writeResponse(untagged, &response{
			text: fmt.Sprintf("SEARCH %v", res.All.String()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *SearchCriteria) And(other *SearchCriteria) {
	c.Before = before(c.Before, other.Before)
	c.Since = after(c.Since, other.Since)
	c.SentBefore = before(c.SentBefore, other.SentBefore)
	c.SentSince = after(c.SentSince, other.SentSince)

	c.Larger = larger(c.Larger, other.Larger)
	c.Smaller = smaller(c.Smaller, other.Smaller)
}

func before(t1, t2 time.Time) time.Time {
	switch {
	case t1.IsZero():
		return t2
	case t2.IsZero():
		return t1
	case t1.Before(t2):
		return t1
	default:
		return t2
	}
}

func after(t1, t2 time.Time) time.Time {
	switch {
	case t1.IsZero():
		return t2
	case t2.IsZero():
		return t1
	case t1.After(t2):
		return t1
	default:
		return t2
	}
}

func larger(v1, v2 int64) int64 {
	if v1 > v2 {
		return v1
	}
	return v2
}

func smaller(v1, v2 int64) int64 {
	if v1 > 0 && v1 < v2 {
		return v1
	}
	return v2
}
