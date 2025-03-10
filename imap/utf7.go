package imap

import (
	"bytes"
	"encoding/base64"
	"unicode/utf16"
)

var base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+,"

func DecodeUTF7(input string) (string, error) {
	var output bytes.Buffer
	var base64Buffer bytes.Buffer
	inBase64 := false

	for i := 0; i < len(input); i++ {
		c := input[i]

		if inBase64 {
			if c == '-' {
				// End of Base64 section
				if base64Buffer.Len() > 0 {
					decoded, err := decodeBase64UTF16(base64Buffer.String())
					if err != nil {
						return "", err
					}
					output.WriteString(decoded)
					base64Buffer.Reset()
				}
				inBase64 = false
			} else {
				base64Buffer.WriteByte(c)
			}
		} else {
			if c == '&' {
				// Start of Base64 section
				inBase64 = true
			} else {
				// Regular ASCII character
				output.WriteByte(c)
			}
		}
	}

	// Handle unclosed Base64 sequences
	if base64Buffer.Len() > 0 {
		decoded, err := decodeBase64UTF16(base64Buffer.String())
		if err != nil {
			return "", err
		}
		output.WriteString(decoded)
	}

	return output.String(), nil
}

func decodeBase64UTF16(encoded string) (string, error) {
	// Fix padding (UTF-7 Base64 does not use standard padding)
	missingPadding := (4 - (len(encoded) % 4)) % 4
	encoded += string(bytes.Repeat([]byte{'='}, missingPadding))

	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	// Convert UTF-16BE bytes to Unicode runes
	var runes []uint16
	for i := 0; i < len(decodedBytes); i += 2 {
		if i+1 < len(decodedBytes) {
			codeUnit := uint16(decodedBytes[i])<<8 | uint16(decodedBytes[i+1])
			runes = append(runes, codeUnit)
		}
	}

	return string(utf16.Decode(runes)), nil
}
