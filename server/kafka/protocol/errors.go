package protocol

type ErrorCode int16

var (
	None                ErrorCode = 0
	IllegalGeneration   ErrorCode = 22
	InvalidGroupId      ErrorCode = 24
	RebalanceInProgress ErrorCode = 27
)
