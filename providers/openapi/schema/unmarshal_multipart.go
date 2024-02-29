package schema

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"mokapi/media"
	"net/http"
	"strings"
)

func readMultipart(data []byte, contentType media.ContentType, r *Ref) (interface{}, error) {
	if r.Value.Type != "object" {
		return nil, fmt.Errorf("schema '%v' not support for content type multipart/form-data, expected 'object'", r.Value.Type)
	}

	if r.Value.Properties == nil {
		// todo raw value
		return nil, nil
	}

	boundary := contentType.Parameters["boundary"]
	reader := multipart.NewReader(strings.NewReader(string(data)), boundary)

	m := make(map[string]interface{})
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := part.FormName()
		p := r.Value.Properties.Get(name)
		v, err := parsePart(part, p)
		if err != nil {
			return nil, err
		}
		m[name] = v
	}
	return m, nil
}

func parsePart(part *multipart.Part, p *Ref) (interface{}, error) {
	defer part.Close()
	b, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	ct := media.ParseContentType(part.Header.Get("Content-Type"))
	return p.Unmarshal(b, ct)
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
