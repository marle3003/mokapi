package kafkatest

import (
	"bytes"
	"mokapi/kafka"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T, version int16, msg kafka.Message) {
	r1 := kafka.Request{
		Header: &kafka.Header{
			ApiKey:        getApiKey(msg),
			ApiVersion:    version,
			ClientId:      "me",
			CorrelationId: 123,
		},
		Message: msg,
	}

	b := &bytes.Buffer{}
	err := r1.Write(b)
	require.NoError(t, err)

	r2 := &kafka.Request{}
	err = r2.Read(b)
	require.NoError(t, err)

	require.True(t, deepEqual(r1.Message, r2.Message))
}

func WriteRequest(t *testing.T, version int16, correlationId int32, clientId string, msg kafka.Message) []byte {
	r := kafka.Request{
		Header: &kafka.Header{
			ApiKey:        getApiKey(msg),
			ApiVersion:    version,
			ClientId:      clientId,
			CorrelationId: correlationId,
		},
		Message: msg,
	}

	b := &bytes.Buffer{}
	err := r.Write(b)
	require.NoError(t, err)
	return b.Bytes()
}

func TestResponse(t *testing.T, version int16, msg kafka.Message) {
	apiKey := getApiKey(msg)
	r1 := kafka.Response{
		Header: &kafka.Header{
			ApiKey:        apiKey,
			ApiVersion:    version,
			CorrelationId: 123,
		},
		Message: msg,
	}

	b := &bytes.Buffer{}
	err := r1.Write(b)
	require.NoError(t, err)

	r2 := &kafka.Response{
		Header: &kafka.Header{
			ApiKey:        apiKey,
			ApiVersion:    version,
			CorrelationId: 123,
		},
	}
	err = r2.Read(b)
	require.NoError(t, err)

	require.True(t, deepEqual(r1.Message, r2.Message))
}

func WriteResponse(t *testing.T, version int16, correlationId int32, msg kafka.Message) []byte {
	r := kafka.Response{
		Header: &kafka.Header{
			ApiKey:        getApiKey(msg),
			ApiVersion:    version,
			CorrelationId: correlationId,
		},
		Message: msg,
	}

	b := &bytes.Buffer{}
	err := r.Write(b)
	require.NoError(t, err)
	return b.Bytes()
}

func deepEqual(i1, i2 any) bool {
	if b1, ok := i1.(kafka.Bytes); ok {
		if b2, ok := i2.(kafka.Bytes); ok {
			if b1.Size() != b2.Size() {
				return false
			}
			return bytes.Equal(kafka.Read(b1), kafka.Read(b2))
		}
		return false
	}

	v1 := reflect.ValueOf(i1)
	v2 := reflect.ValueOf(i2)

	t1 := reflect.TypeOf(i1)
	if v1.Type() != v2.Type() {
		return false
	}

	switch v1.Kind() {
	case reflect.Struct:
		for i := 0; i < v1.NumField(); i++ {
			if !t1.Field(i).IsExported() {
				continue
			}
			if !deepEqual(v1.Field(i).Interface(), v2.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Ptr:
		if v1.IsNil() {
			return v2.IsNil()
		}
		return deepEqual(v1.Elem().Interface(), v2.Elem().Interface())
	case reflect.Slice:
		if v1.Len() != v2.Len() {
			return false
		}
		for i := 0; i < v1.Len(); i++ {
			if !deepEqual(v1.Index(i).Interface(), v2.Index(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Map:
		if v1.Len() != v2.Len() {
			return false
		}
		for _, k := range v1.MapKeys() {
			if !deepEqual(v1.MapIndex(k).Interface(), v2.MapIndex(k).Interface()) {
				return false
			}
		}
		return true
	default:
		return i1 == i2
	}
}
