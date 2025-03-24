package imap

import (
	"fmt"
	"strings"
)

type Set interface {
	Contains(num uint32) bool
	String() string
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

func (n IdNum) Contains(num uint32) bool {
	return uint32(n) == num
}

func (n IdNum) String() string {
	return fmt.Sprintf("%d", n)
}
