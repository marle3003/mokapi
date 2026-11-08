package mail_test

import (
	"encoding/json"
	"mokapi/engine/common"
	"mokapi/providers/mail"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLog_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		log  func() *mail.Log
		json string
	}{
		{
			name: "empty",
			log: func() *mail.Log {
				return &mail.Log{}
			},
			json: `{"from":"","to":null,"messageId":"","subject":"","duration":0,"error":"","actions":null}`,
		},
		{
			name: "with duration and actions",
			log: func() *mail.Log {
				l := &mail.Log{
					From:      "alice@foo.bar",
					To:        []string{"bob@foo.bar"},
					MessageId: "id",
					Subject:   "Hello",
					Error:     "failed",
				}
				l.SetDuration(12)
				l.SetActions([]*common.Action{{Duration: 2}})
				return l
			},
			json: `{"from":"alice@foo.bar","to":["bob@foo.bar"],"messageId":"id","subject":"Hello","duration":12,"error":"failed","actions":[{"duration":2,"tags":null,"parameters":null,"logs":null,"error":null}]}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.log())
			require.NoError(t, err)
			require.Equal(t, tc.json, string(b))
		})
	}
}
