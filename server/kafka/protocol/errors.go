package protocol

type ErrorCode int16

var (
	None                ErrorCode = 0
	InvalidGroupId      ErrorCode = 24
	RebalanceInProgress ErrorCode = 27
)
