package smtptest

import (
	"bufio"
	"io"
)

type Response struct {
	Message string
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) Read(reader io.Reader) error {
	br := bufio.NewReader(reader)
	line, _, err := br.ReadLine()
	r.Message = string(line)
	return err
}
