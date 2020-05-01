package handlers

import "net/http"

type Context struct {
	Response   http.ResponseWriter
	Request    *http.Request
	ServiceUrl string
}

func NewContext(serviceUrl string, response http.ResponseWriter, request *http.Request) *Context {
	return &Context{Response: response, Request: request, ServiceUrl: serviceUrl}
}
