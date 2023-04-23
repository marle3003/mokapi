package kafka

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type ApiKey int16

const (
	Produce         ApiKey = 0
	Fetch           ApiKey = 1
	Offset          ApiKey = 2
	Metadata        ApiKey = 3
	OffsetCommit    ApiKey = 8
	OffsetFetch     ApiKey = 9
	FindCoordinator ApiKey = 10
	JoinGroup       ApiKey = 11
	Heartbeat       ApiKey = 12
	SyncGroup       ApiKey = 14
	ListGroup       ApiKey = 16
	ApiVersions     ApiKey = 18
	CreateTopics    ApiKey = 19
)

var apitext = map[ApiKey]string{
	Produce:         "Produce",
	Fetch:           "Fetch",
	Offset:          "Offset",
	Metadata:        "Metadata",
	OffsetCommit:    "OffsetCommit",
	OffsetFetch:     "OffsetFetch",
	FindCoordinator: "FindCoordinator",
	JoinGroup:       "JoinGroup",
	Heartbeat:       "Heartbeat",
	SyncGroup:       "SyncGroup",
	ApiVersions:     "ApiVersions",
	CreateTopics:    "CreateTopics",
}

var ApiTypes = map[ApiKey]ApiType{}

type Message interface {
}

type ApiReg struct {
	ApiKey     ApiKey
	MinVersion int16
	MaxVersion int16
}

type decodeMsg func(*Decoder, int16) (Message, error)
type encodeMsg func(*Encoder, int16, Message) error

type encoding struct {
	decode decodeMsg
	encode encodeMsg
}

type messageType struct {
	decode map[int16]decodeFunc
	encode map[int16]encodeFunc
}

type ApiType struct {
	MinVersion int16
	MaxVersion int16
	request    encoding
	response   encoding
	// https://cwiki.apache.org/confluence/display/KAFKA/KIP-482%3A+The+Kafka+Protocol+should+Support+Optional+Tagged+Fields
	flexibleRequest  int16
	flexibleResponse int16
}

type kafkaTag struct {
	minVersion int16
	maxVersion int16
	// version to switch to compact mode (inclusive)
	compact   int16
	protoType string
	nullable  bool
}

func (t kafkaTag) isValid(version int16) bool {
	return t.minVersion <= version && t.maxVersion >= version
}

type Header struct {
	Size          int32
	ApiKey        ApiKey
	ApiVersion    int16
	CorrelationId int32
	ClientId      string           `kafka:"type=NULLABLE_STRING"`
	TagFields     map[int64]string `kafka:"type=TAG_BUFFER"`
}

func Register(reg ApiReg, req, res Message, flexibleRequest int16, flexibleResponse int16) {
	tReq := reflect.ValueOf(req).Elem().Type()
	tRes := reflect.ValueOf(res).Elem().Type()

	requestTypes := newMessageType()
	responseTypes := newMessageType()
	tag := kafkaTag{}

	for i := reg.MinVersion; i <= reg.MaxVersion; i++ {
		requestTypes.decode[i] = newDecodeFunc(tReq, i, tag)
		requestTypes.encode[i] = newEncodeFunc(tReq, i, tag)

		responseTypes.decode[i] = newDecodeFunc(tRes, i, tag)
		responseTypes.encode[i] = newEncodeFunc(tRes, i, tag)
	}

	ApiTypes[reg.ApiKey] = ApiType{
		reg.MinVersion,
		reg.MaxVersion,
		encoding{
			func(d *Decoder, version int16) (Message, error) {
				decode, ok := requestTypes.decode[version]
				if !ok {
					return nil, fmt.Errorf("unsupported version %v", version)
				}
				msg := reflect.New(tReq).Interface().(Message)
				decode(d, reflect.ValueOf(msg).Elem())
				return msg, nil
			},
			func(e *Encoder, version int16, msg Message) error {
				encode, ok := requestTypes.encode[version]
				if !ok {
					return fmt.Errorf("unsupported version %v", version)
				}
				encode(e, reflect.ValueOf(msg).Elem())
				return nil
			},
		},
		encoding{
			func(d *Decoder, version int16) (Message, error) {
				decode, ok := responseTypes.decode[version]
				if !ok {
					return nil, fmt.Errorf("unsupported version %v", version)
				}
				msg := reflect.New(tRes).Interface().(Message)
				decode(d, reflect.ValueOf(msg).Elem())
				return msg, nil
			},
			func(e *Encoder, version int16, msg Message) error {
				encode, ok := responseTypes.encode[version]
				if !ok {
					return fmt.Errorf("unsupported version %v", version)
				}
				encode(e, reflect.ValueOf(msg).Elem())
				return nil
			},
		},
		flexibleRequest,
		flexibleResponse,
	}
}

func getTag(f reflect.StructField) kafkaTag {
	if s, ok := f.Tag.Lookup("kafka"); !ok {
		return kafkaTag{}
	} else {
		s := strings.Split(s, ",")
		t := kafkaTag{minVersion: 0, maxVersion: math.MaxInt16, compact: math.MaxInt16}
		for _, opt := range s {
			opt = strings.TrimSpace(opt)
			kv := strings.Split(opt, "=")
			switch kv[0] {
			case "min":
				if i, err := strconv.Atoi(kv[1]); err == nil {
					t.minVersion = int16(i)
				}
			case "max":
				if i, err := strconv.Atoi(kv[1]); err == nil {
					t.maxVersion = int16(i)
				}
			case "compact":
				if i, err := strconv.Atoi(kv[1]); err == nil {
					t.compact = int16(i)
				}
			case "type":
				t.protoType = kv[1]
			case "nullable":
				if len(kv) == 1 {
					t.nullable = true
				} // else: parse bool value
			}
		}
		return t
	}
}

func readHeader(d *Decoder) (h *Header) {
	h = &Header{}

	h.Size = d.ReadInt32()
	if h.Size == 0 {
		return
	}

	d.leftSize = int(h.Size)

	h.ApiKey = ApiKey(d.ReadInt16())
	h.ApiVersion = d.ReadInt16()
	h.CorrelationId = d.ReadInt32()
	h.ClientId = d.ReadString()

	if h.ApiVersion >= ApiTypes[h.ApiKey].flexibleRequest {
		h.TagFields = d.ReadTagFields()
	}

	return
}

func newMessageType() *messageType {
	return &messageType{
		decode: make(map[int16]decodeFunc),
		encode: make(map[int16]encodeFunc),
	}
}

func (a ApiKey) String() string {
	if s, ok := apitext[a]; ok {
		return fmt.Sprintf("%v (%v)", s, int(a))
	}

	return fmt.Sprintf("unknown kafka api key: %v", int(a))
}
