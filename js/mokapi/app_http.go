package mokapi

import (
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"strings"

	"github.com/dop251/goja"
)

type Http struct {
	filter common.HttpFilter
	m      *Module
}

type HttpRoute struct {
	filter common.HttpFilter
	m      *Module
}

func (h *Http) Route(path string) goja.Value {
	filter := h.filter
	filter.Path = path
	r := &HttpRoute{filter: filter, m: h.m}
	return newRouteObject(r)
}

func (h *Http) Operation(operationId string) goja.Value {
	filter := h.filter
	filter.OperationId = operationId
	r := &HttpRoute{filter: filter, m: h.m}
	return newRouteObject(r)
}

func (h *Http) Method(method string, path string, do goja.Value, vArgs goja.Value) {
	filter := h.filter
	filter.Path = path
	filter.Method = method

	args, err := getOnArgs(h.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, h.m.vm, h.m.loop)

	h.m.host.OnHttp(filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
}

func (h *Http) Api(name string) goja.Value {
	f := h.filter
	f.Api = name
	return newHttpObject(&Http{filter: f, m: h.m})
}

func (h *Http) Use(do goja.Value, vArgs goja.Value) goja.Value {
	args, err := getOnArgs(h.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, h.m.vm, h.m.loop)

	h.m.host.OnHttp(h.filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return newHttpObject(h)
}

func (h *HttpRoute) Method(method string, do goja.Value, vArgs goja.Value) {
	filter := h.filter
	filter.Method = method

	args, err := getOnArgs(h.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, h.m.vm, h.m.loop)

	h.m.host.OnHttp(filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
}

func (h *HttpRoute) Use(do goja.Value, vArgs goja.Value) goja.Value {
	args, err := getOnArgs(h.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	f := getHandler(do, args, h.m.vm, h.m.loop)

	h.m.host.OnHttp(h.filter, f, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return newRouteObject(h)
}

type HttpObject struct {
	http *Http
	vm   *goja.Runtime
	loop *eventloop.EventLoop
}

func newHttpObject(h *Http) goja.Value {
	do := h.m.vm.NewDynamicObject(&HttpObject{
		http: h,
		vm:   h.m.vm,
		loop: h.m.loop,
	})
	return do
}

func (r *HttpObject) Get(key string) goja.Value {
	switch key {
	case "use":
		return r.vm.ToValue(r.http.Use)
	case "route":
		return r.vm.ToValue(r.http.Route)
	case "api":
		return r.vm.ToValue(r.http.Api)
	}

	method := strings.ToUpper(key)
	return r.vm.ToValue(func(path string, do goja.Value, vArgs goja.Value) {
		r.http.Method(method, path, do, vArgs)
	})
}

func (r *HttpObject) Set(_ string, _ goja.Value) bool { return false }
func (r *HttpObject) Has(_ string) bool               { return true }
func (r *HttpObject) Delete(_ string) bool            { return false }
func (r *HttpObject) Keys() []string {
	return []string{"use", "api", "get", "post", "put", "patch", "delete", "head", "options"}
}

type RouteObject struct {
	route *HttpRoute
	vm    *goja.Runtime
	loop  *eventloop.EventLoop
}

func newRouteObject(route *HttpRoute) goja.Value {
	do := route.m.vm.NewDynamicObject(&RouteObject{
		route: route,
		vm:    route.m.vm,
		loop:  route.m.loop,
	})
	return do
}

func (r *RouteObject) Get(key string) goja.Value {
	switch key {
	case "use":
		return r.vm.ToValue(r.route.Use)
	}

	method := strings.ToUpper(key)
	return r.vm.ToValue(func(do goja.Value, vArgs goja.Value) {
		r.route.Method(method, do, vArgs)
	})
}

func (r *RouteObject) Set(_ string, _ goja.Value) bool { return false }
func (r *RouteObject) Has(_ string) bool               { return true }
func (r *RouteObject) Delete(_ string) bool            { return false }
func (r *RouteObject) Keys() []string {
	return []string{"use", "get", "post", "put", "patch", "delete", "head", "options"}
}
