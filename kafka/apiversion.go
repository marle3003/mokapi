package kafka

import (
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"sort"
)

func (b *Broker) apiversion(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*apiVersion.Request)

	if req.Header.ApiVersion >= 3 {
		client := req.Context.Value("client").(*ClientContext)
		client.clientSoftwareName = r.ClientSwName
		client.clientSoftwareVersion = r.ClientSwVersion
	}

	res := &apiVersion.Response{
		ApiKeys: make([]apiVersion.ApiKeyResponse, 0, len(protocol.ApiTypes)),
	}
	keys := make([]int, 0, len(protocol.ApiTypes))
	for k := range protocol.ApiTypes {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		key := protocol.ApiKey(k)
		t := protocol.ApiTypes[key]
		res.ApiKeys = append(res.ApiKeys, apiVersion.NewApiKeyResponse(key, t))
	}
	return rw.Write(res)
}
