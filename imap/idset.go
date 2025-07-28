package imap

import (
	"fmt"
	"strings"
)

type Set interface {
	Contains(num uint32) bool
	String() string
	Nums() ([]uint32, bool)
}

type IdSet struct {
	Ids   []Set
	IsUid bool
}

type Range struct {
	Start SeqNum
	End   SeqNum
}

type IdNum uint32

type SeqNum struct {
	Value uint32
	Star  bool
}

func (s *IdSet) Contains(num uint32) bool {
	for _, set := range s.Ids {
		if set.Contains(num) {
			return true
		}
	}
	return false
}

func (s *IdSet) String() string {
	var sb strings.Builder
	for _, set := range s.Ids {
		if sb.Len() != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(set.String())
	}
	return sb.String()
}

func (s *IdSet) Append(v Set) {
	s.Ids = append(s.Ids, v)
}

func (s *IdSet) AddId(num uint32) {
	s.Ids = append(s.Ids, IdNum(num))
}

func (s *IdSet) AddRange(start, end SeqNum) {
	s.Ids = append(s.Ids, &Range{Start: start, End: end})
}

func (s *IdSet) Nums() ([]uint32, bool) {
	var results []uint32
	for _, set := range s.Ids {
		nums, b := set.Nums()
		if !b {
			return nil, false
		}
		results = append(results, nums...)
	}
	return results, true
}

func (s *Range) Contains(num uint32) bool {
	if num < s.Start.Value {
		return false
	}
	if s.End.Star {
		return true
	} else {
		return s.Start.Value <= num && s.End.Value >= num
	}
}

func (s *Range) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d", s.Start.Value))
	sb.WriteByte(':')
	if s.End.Star {
		sb.WriteString("*")
	} else {
		sb.WriteString(fmt.Sprintf("%d", s.End.Value))
	}
	return sb.String()
}

func (s *Range) Nums() ([]uint32, bool) {
	if s.Start.Star || s.End.Star {
		return nil, false
	}
	var results []uint32
	for i := s.Start.Value; i <= s.End.Value; i++ {
		results = append(results, i)
	}
	return results, true
}

func (n IdNum) Contains(num uint32) bool {
	return uint32(n) == num
}

func (n IdNum) String() string {
	return fmt.Sprintf("%d", n)
}

func (n IdNum) Nums() ([]uint32, bool) {
	return []uint32{uint32(n)}, true
}

func parseSequence(s string) (IdSet, error) {
	set := IdSet{}

	var err error
	for _, v := range strings.Split(s, ",") {
		if i := strings.IndexRune(v, ':'); i >= 0 {
			r := &Range{}
			r.Start, err = parseNumSet(v[:i])
			if err != nil {
				return set, err
			}
			r.End, err = parseNumSet(v[i+1:])
			if err != nil {
				return set, err
			}
			set.Ids = append(set.Ids, r)
		} else {
			var n SeqNum
			n, err = parseNumSet(v)
			if err != nil {
				return set, err
			}
			set.Ids = append(set.Ids, IdNum(n.Value))
		}
	}
	return set, nil
}
