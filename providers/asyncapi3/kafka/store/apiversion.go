package store

import (
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"sort"
)

func (s *Store) apiversion(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*apiVersion.Request)

	res := &apiVersion.Response{
		ApiKeys: make([]apiVersion.ApiKeyResponse, 0, len(kafka.ApiTypes)),
	}

	if req.Header.ApiVersion > kafka.ApiTypes[kafka.ApiVersions].MaxVersion {
		// client is ahead, broker must respond with version 0 and error code
		// UNSUPPORTED_VERSION
		req.Header.ApiVersion = 0
		res.ErrorCode = kafka.UnsupportedVersion
	}

	if req.Header.ApiVersion >= 3 {
		client := kafka.ClientFromContext(req)
		client.ClientSoftwareName = r.ClientSwName
		client.ClientSoftwareVersion = r.ClientSwVersion
	}

	keys := make([]int, 0, len(kafka.ApiTypes))
	for k := range kafka.ApiTypes {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		key := kafka.ApiKey(k)
		t := kafka.ApiTypes[key]
		res.ApiKeys = append(res.ApiKeys, apiVersion.NewApiKeyResponse(key, t))
	}
	return rw.Write(res)
}
