// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

type Route struct {
	Name, method string
	handlers     []Handler
	tags         []Tag
	above, below *Router
	index        int
}

type Router struct {
	routes   []*Route
	handlers []Handler
	above    *Route
}

func (r *Router) Fallback() {
	r.Use(
		HandleNotImplemented,
		NewRedirectTrailingSlash(true),
		NewRedirectCleanedPath(true),
		NewAllowedMethods(true),
		// HandleNotFound,
	)
}

func (r *Router) Use(handlers ...interface{}) {
	r.handlers = append(r.handlers, newHandlers(handlers)...)
}

func (r *Router) Group(path string, f func(*Router), handlers ...interface{}) {
	rr := &Route{
		tags:     mustSplitPath(path),
		handlers: newHandlers(handlers),
		above:    r,
		below:    new(Router),
	}
	rr.below.above = rr
	r.routes = append(r.routes, rr)
	rr.index = len(r.routes)
	f(rr.below)
}

func (r *Router) Handle(method, path string, handlers ...interface{}) *Route {
	rr := &Route{
		method:   method,
		tags:     mustSplitPath(path),
		handlers: newHandlers(handlers),
		above:    r,
	}
	r.routes = append(r.routes, rr)
	return rr
}

func (r *Router) Any(path string, handlers ...interface{}) *Route {
	return r.Handle("", path, handlers...)
}

func (r *Router) CONNECT(path string, handlers ...interface{}) *Route {
	return r.Handle("CONNECT", path, handlers...)
}

func (r *Router) DELETE(path string, handlers ...interface{}) *Route {
	return r.Handle("DELETE", path, handlers...)
}

func (r *Router) GET(path string, handlers ...interface{}) *Route {
	return r.Handle("GET", path, handlers...)
}

func (r *Router) HEAD(path string, handlers ...interface{}) *Route {
	return r.Handle("HEAD", path, handlers...)
}

func (r *Router) OPTIONS(path string, handlers ...interface{}) *Route {
	return r.Handle("OPTIONS", path, handlers...)
}

func (r *Router) POST(path string, handlers ...interface{}) *Route {
	return r.Handle("POST", path, handlers...)
}

func (r *Router) PUT(path string, handlers ...interface{}) *Route {
	return r.Handle("PUT", path, handlers...)
}

func (r *Router) TRACE(path string, handlers ...interface{}) *Route {
	return r.Handle("TRACE", path, handlers...)
}

var DefaultRouter = new(Router)

func Use(handlers ...interface{}) {
	DefaultRouter.Use(handlers...)
}

func Group(pattern string, f func(*Router), handlers ...interface{}) {
	DefaultRouter.Group(pattern, f, handlers...)
}

func Handle(method, pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.Handle(method, pattern, handlers...)
}

func Any(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.Any(pattern, handlers...)
}

func GET(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.GET(pattern, handlers...)
}

func HEAD(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.HEAD(pattern, handlers...)
}

func POST(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.POST(pattern, handlers...)
}

func PUT(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.PUT(pattern, handlers...)
}

func DELETE(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.DELETE(pattern, handlers...)
}

func CONNECT(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.CONNECT(pattern, handlers...)
}

func OPTIONS(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.OPTIONS(pattern, handlers...)
}

func TRACE(pattern string, handlers ...interface{}) *Route {
	return DefaultRouter.TRACE(pattern, handlers...)
}
