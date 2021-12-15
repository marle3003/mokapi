package kafka

import (
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
)

func (b *BrokerServer) apiversion(rw protocol.ResponseWriter, req *protocol.Request) error {
	r := req.Message.(*apiVersion.Request)

	if req.Header.ApiVersion >= 3 {
		req.Context.WithValue("ClientSoftwareName", r.ClientSwName)
		req.Context.WithValue("ClientSoftwareVersion", r.ClientSwVersion)
	}

	res := &apiVersion.Response{
		ApiKeys: make([]apiVersion.ApiKeyResponse, 0, len(protocol.ApiTypes)),
	}
	for k, t := range protocol.ApiTypes {
		res.ApiKeys = append(res.ApiKeys, apiVersion.NewApiKeyResponse(k, t))
	}
	return rw.Write(res)
}
