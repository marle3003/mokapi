package protocol

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type ApiKey int16

const (
	Produce         ApiKey = 0
	Fetch           ApiKey = 1
	ListOffsets     ApiKey = 2
	Metadata        ApiKey = 3
	OffsetCommit    ApiKey = 8
	OffsetFetch     ApiKey = 9
	FindCoordinator ApiKey = 10
	JoinGroup       ApiKey = 11
	Heartbeat       ApiKey = 12
	SyncGroup       ApiKey = 14
	ApiVersions     ApiKey = 18
)

const (
	UnknownTopicOrPartition = 3
)

var (
	pagePool = sync.Pool{New: func() interface{} { return new(page) }}
)

var ApiTypes = map[ApiKey]ApiType{}

type Message interface {
}

type ApiReg struct {
	ApiKey     ApiKey
	MinVersion int16
	MaxVersion int16
}

type decodeMsg func(*Decoder, int16) Message
type encodeMsg func(*Encoder, int16, Message)

type ApiType struct {
	MinVersion int16
	MaxVersion int16
	decode     decodeMsg
	encode     encodeMsg
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
	val := reflect.ValueOf(req).Elem()
	t := val.Type()

	decode := make(map[int16]decodeFunc)
	encode := make(map[int16]encodeFunc)
	tag := kafkaTag{}

	for i := reg.MinVersion; i <= reg.MaxVersion; i++ {
		decode[i] = newDecodeFunc(t, i, tag)
		encode[i] = newEncodeFunc(reflect.ValueOf(res).Elem().Type(), i, tag)
	}

	ApiTypes[reg.ApiKey] = ApiType{
		reg.MinVersion,
		reg.MaxVersion,
		func(d *Decoder, version int16) Message {
			msg := reflect.New(t).Interface().(Message)
			decode[version](d, reflect.ValueOf(msg).Elem())
			return msg
		},
		func(e *Encoder, version int16, msg Message) {
			encode[version](e, reflect.ValueOf(msg).Elem())
		},
		flexibleRequest,
		flexibleResponse,
	}
}

func ReadMessage(r io.Reader) (h *Header, msg Message, err error) {
	d := NewDecoder(r, 4)
	h = readHeader(d)

	if h.Size == 0 {
		return nil, nil, io.EOF
	}

	if d.err != nil {
		err = d.err
		return
	}

	t := ApiTypes[h.ApiKey]
	msg = t.decode(d, h.ApiVersion)
	err = d.err

	return
}

func WriteMessage(w io.Writer, k ApiKey, version int16, correlationId int32, msg Message) {
	p := pagePool.Get().(*page)
	defer func() {
		p.Reset()
		pagePool.Put(p)
	}()

	e := NewEncoder(p)
	t := ApiTypes[k]

	e.writeInt32(0)
	e.writeInt32(correlationId)
	if version >= t.flexibleResponse {
		e.writeUVarInt(0) // tag_buffer
	}
	t.encode(e, version, msg)

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(p.Size()-4))
	p.WriteAt(size[:], 0)

	_, err := w.Write(p.buffer[0:p.offset])
	if err != nil {
		log.Errorf("unable to write kafka message apikey %q", k)
	}
}

func getTag(f reflect.StructField) kafkaTag {
	if s, ok := f.Tag.Lookup("kafka"); !ok {
		return kafkaTag{}
	} else {
		s := strings.Split(s, ",")
		t := kafkaTag{minVersion: 0, maxVersion: math.MaxInt16, compact: math.MaxInt16}
		for _, opt := range s {
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

	h.Size = d.readInt32()
	if h.Size == 0 {
		return
	}

	d.leftSize = int(h.Size)

	h.ApiKey = ApiKey(d.readInt16())
	h.ApiVersion = d.readInt16()
	h.CorrelationId = d.readInt32()
	h.ClientId = d.readString()

	if h.ApiVersion >= ApiTypes[h.ApiKey].flexibleRequest {
		h.TagFields = d.readTagFields()
	}

	return
}
