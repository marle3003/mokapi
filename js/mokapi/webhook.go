package mokapi

import (
	"mokapi/engine/common"
)

func (m *Module) Webhook(name, url string, args common.WebhookArgs) *common.WebhookResponse {
	if name == "" {
		panic(m.vm.ToValue("webhook name must not be empty"))
	}
	if url == "" {
		panic(m.vm.ToValue("webhook url must not be empty"))
	}

	res, err := m.host.Webhook(name, url, args)
	if err != nil {
		panic(err)
	}
	return res
}
