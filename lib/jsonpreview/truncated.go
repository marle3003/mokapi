package jsonpreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"
)

type Truncated struct {
	V     any
	Limit int
}

func (t Truncated) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(t.V)
	if err != nil {
		return nil, err
	}
	if len(data) <= t.Limit {
		return data, nil
	}
	cut := t.Limit
	// do not cut in the middle of a UTF-8 char
	for cut > 0 && !utf8.RuneStart(data[cut]) {
		cut--
	}
	sum := sha256.Sum256(data)
	return json.Marshal(struct {
		Truncated bool   `json:"truncated"`
		Size      int    `json:"size"`
		Hash      string `json:"hash"`
		Preview   string `json:"preview"`
	}{
		Truncated: true,
		Size:      len(data),
		Hash:      hex.EncodeToString(sum[:]),
		Preview:   string(data[:cut]),
	})
}
