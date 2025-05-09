package encoding

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mime/multipart"
	"mokapi/media"
	"net/http"
)

type BinaryDecoder struct {
}

func (d *BinaryDecoder) IsSupporting(contentType media.ContentType) bool {
	return contentType.String() == "application/octet-stream"
}

func (d *BinaryDecoder) Decode(b []byte, state *DecodeState) (i interface{}, err error) {
	return state.parser.Parse(b)
}

func parseFormFile(fh *multipart.FileHeader) (interface{}, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Errorf("unable to close file: %v", err)
		}
	}()

	var sniff [512]byte
	_, err = f.Read(sniff[:])
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"filename": fh.Filename,
		"type":     http.DetectContentType(sniff[:]),
		"size":     prettyByteCountIEC(fh.Size),
	}, nil
}

func prettyByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
