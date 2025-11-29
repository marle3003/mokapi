package initProducerId

import "mokapi/kafka"

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.InitProducerId,
			MinVersion: 0,
			MaxVersion: 6,
		},
		&Request{},
		&Response{},
		3,
		3,
	)
}

type Request struct {
	TransactionalId      string `kafka:"nullable,compact=2"`
	TransactionTimeoutMs int32  `kafka:""`
	ProducerId           int64  `kafka:"min=3"`
	ProducerEpoch        int16  `kafka:"min=3"`
	// Enable2PC true if the client wants to enable two-phase commit (2PC) for transaction
	Enable2PC bool             `kafka:"min=6"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type Response struct {
	ThrottleTimeMs          int32            `kafka:""`
	ErrorCode               kafka.ErrorCode  `kafka:""`
	ProducerId              int64            `kafka:""`
	ProducerEpoch           int16            `kafka:""`
	OngoingTxnProducerId    int64            `kafka:"min=6"`
	OngoingTxnProducerEpoch int16            `kafka:"min=6"`
	TagFields               map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}
