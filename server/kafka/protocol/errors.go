package protocol

import "fmt"

type ErrorCode int16

const (
	None                    ErrorCode = 0
	OffsetOutOfRange        ErrorCode = 1
	UnknownTopicOrPartition ErrorCode = 3
	InvalidTopic            ErrorCode = 17
	IllegalGeneration       ErrorCode = 22
	InvalidGroupId          ErrorCode = 24
	UnknownMemberId         ErrorCode = 25
	RebalanceInProgress     ErrorCode = 27
	GroupIdNotFound         ErrorCode = 69
	MemberIdRequired        ErrorCode = 79
)

var (
	errorCodeText = map[ErrorCode]string{
		None:                    "NONE",
		OffsetOutOfRange:        "OFFSET_OUT_OF_RANGE",
		UnknownTopicOrPartition: "UNKNOWN_TOPIC_OR_PARTITION",
		InvalidTopic:            "INVALID_TOPIC_EXCEPTION",
		IllegalGeneration:       "ILLEGAL_GENERATION",
		InvalidGroupId:          "INVALID_GROUP_ID",
		UnknownMemberId:         "UNKNOWN_MEMBER_ID",
		RebalanceInProgress:     "REBALANCE_IN_PROGRESS",
		GroupIdNotFound:         "GROUP_ID_NOT_FOUND",
		MemberIdRequired:        "MEMBER_ID_REQUIRED",
	}
)

func (e ErrorCode) String() string {
	if s, ok := errorCodeText[e]; ok {
		return fmt.Sprintf("%v (%v)", s, int(e))
	}

	return fmt.Sprintf("unknown kafka error code: %v", int(e))
}
