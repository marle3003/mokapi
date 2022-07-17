package smtp

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	mail "mokapi/server/smtp"
)

type Module struct {
	host common.Host
	rt   *goja.Runtime
}

type mailArgs struct {
	Address string
}

func New(host common.Host, rt *goja.Runtime) interface{} {
	return &Module{host: host, rt: rt}
}

func (m *Module) Send(msgV goja.Value, argsV goja.Value) error {
	mArgs := &mailArgs{}
	if argsV != nil && !goja.IsUndefined(argsV) && !goja.IsNull(argsV) {
		params := argsV.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "address":
				addressV := params.Get(k)
				mArgs.Address = addressV.String()
			}
		}
	}

	msg := msgV.Export().(*mail.Mail)
	to := make([]string, 0, len(msg.To))
	for _, v := range msg.To {
		to = append(to, v.Address)
	}

	//mail.
	//	smtp.SendMail(mArgs.Address, nil, msg.Sender.Address, to)

	return nil
}
