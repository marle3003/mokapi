package lua

import "mokapi/engine/common"

type file struct {
	host common.Host
}

func newFile(host common.Host) *file {
	return &file{host: host}
}

func (f *file) open(filename string) (string, string) {
	s, err := f.host.OpenFile(filename)
	if err != nil {
		return "", err.Error()
	}
	return s, ""
}
