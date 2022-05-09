package metrics

import (
	"bytes"
	"encoding/json"
	"strings"
)

func writeAttr(b *bytes.Buffer, label string, value interface{}) {
	label = strings.ReplaceAll(label, " ", "_")
	key, _ := json.Marshal(label)
	b.Write(key)
	b.WriteRune(':')
	v, _ := json.Marshal(value)
	b.Write(v)
}
