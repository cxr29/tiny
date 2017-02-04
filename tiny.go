// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"net"
	"net/http"
	"net/http/fcgi"
)

type Handler interface {
	ServeHTTP(*Context)
}

type HandlerFunc func(*Context)

func (f HandlerFunc) ServeHTTP(ctx *Context) {
	f(ctx)
}

func newHandler(handler interface{}) Handler {
	switch h := handler.(type) {
	case Handler:
		return h
	case http.Handler:
		return HandlerFunc(func(ctx *Context) {
			h.ServeHTTP(ctx, ctx.Request)
		})
	case func(*Context):
		return HandlerFunc(h)
	case func(http.ResponseWriter, *http.Request):
		return HandlerFunc(func(ctx *Context) {
			h(ctx, ctx.Request)
		})
	}
	panic("unsupported handler")
}

func newHandlers(handlers []interface{}) (a []Handler) {
	if len(handlers) > 0 {
		a = make([]Handler, len(handlers))
		for i, h := range handlers {
			a[i] = newHandler(h)
		}
	}
	return
}

func copyHandlers(handlers []Handler) (a []Handler) {
	if len(handlers) > 0 {
		a = make([]Handler, len(handlers))
		copy(a, handlers)
	}
	return
}

func (r *Router) Handler() http.Handler {
	return newTree(r)
}

func ListenAndServe(addr string, handler http.Handler) error {
	if handler == nil {
		handler = DefaultRouter.Handler()
	}
	return http.ListenAndServe(addr, handler)
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	if handler == nil {
		handler = DefaultRouter.Handler()
	}
	return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
}

func ListenAndServeFCGI(addr string, handler http.Handler) error {
	if handler == nil {
		handler = DefaultRouter.Handler()
	}
	if addr == "" {
		return fcgi.Serve(nil, handler)
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	return fcgi.Serve(l, handler)
}
