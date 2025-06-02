package store

import (
	"mokapi/mqtt"
)

func (s *Store) connect(rw mqtt.ResponseWriter, r *mqtt.Request) {
	connect := r.Message.(*mqtt.ConnectRequest)

	if len(connect.ClientId) == 0 || len(connect.ClientId) > 23 {
		rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
			SessionPresent: false,
			ReturnCode:     mqtt.ErrIdentifierRejected,
		})
		return
	}
}
