package protocol

type ErrorCode int16

var (
	None                ErrorCode = 0
	OffsetOutOfRange    ErrorCode = 1
	IllegalGeneration   ErrorCode = 22
	InvalidGroupId      ErrorCode = 24
	RebalanceInProgress ErrorCode = 27
	GroupIdNotFound     ErrorCode = 69
)
