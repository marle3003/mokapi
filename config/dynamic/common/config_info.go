package common

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
	"time"
)

type ConfigInfo struct {
	Provider string
	Url      *url.URL
	Checksum []byte
	Time     time.Time
	inner    *ConfigInfo
}

func (ci *ConfigInfo) Path() string {
	if len(ci.Url.Opaque) > 0 {
		return ci.Url.Opaque
	}
	u := ci.Url
	path, _ := url.PathUnescape(ci.Url.Path)
	query, _ := url.QueryUnescape(ci.Url.RawQuery)
	var sb strings.Builder
	if len(u.Scheme) > 0 {
		sb.WriteString(u.Scheme + ":")
	}
	if len(u.Scheme) > 0 || len(u.Host) > 0 {
		sb.WriteString("//")
	}
	if len(u.Host) > 0 {
		sb.WriteString(u.Host)
	}
	sb.WriteString(path)
	if len(query) > 0 {
		sb.WriteString("?" + query)
	}
	return sb.String()
}

func (ci *ConfigInfo) Inner() *ConfigInfo {
	return ci.inner
}

func (ci *ConfigInfo) Update(checksum []byte) {
	ci.Time = time.Now()
	ci.Checksum = checksum
}

func (ci *ConfigInfo) Key() string {
	hash := md5.New()
	hash.Write([]byte(ci.Url.String()))
	s := hex.EncodeToString(hash.Sum(nil))
	id, err := uuid.FromBytes([]byte(s[0:16]))
	if err != nil {
		log.Errorf("generate config key '%v' failed: %v", ci.Url.String(), err)
	}
	return id.String()
}
