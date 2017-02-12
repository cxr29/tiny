// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"net/http"
	"path"
	"sort"
	"strings"
)

func HandleNotImplemented(ctx *Context) {
	if ctx.Routed() || ctx.Request.Method == "OPTIONS" {
		return
	}
	if _, ok := ctx.tree.methods[ctx.Request.Method]; !ok {
		ctx.errorStatus(http.StatusNotImplemented)
	}
}

func hasTrailingSlash(p string) bool {
	return len(p) > 0 && p[len(p)-1] == '/'
}

func redirectStatusCode(method string, permanent bool) int {
	if method == "GET" {
		if permanent {
			return http.StatusMovedPermanently
		}
		return http.StatusFound
	}
	if permanent {
		return http.StatusPermanentRedirect
	}
	return http.StatusTemporaryRedirect
}

func redirect(ctx *Context, p string, permanent bool) {
	n, params := ctx.tree.match(ctx.Request.Method, p)
	ctx.tree.putParams(params)
	if n != nil {
		if p == "" {
			ctx.Request.URL.Path = p
			ctx.node = n // perfect
		} else {
			u := *ctx.Request.URL
			u.Path = p
			http.Redirect(ctx, ctx.Request, u.String(), redirectStatusCode(ctx.Request.Method, permanent))
		}
	}
}

func NewRedirectTrailingSlash(permanent bool) Handler {
	return HandlerFunc(func(ctx *Context) {
		if ctx.Routed() || ctx.Request.Method == "CONNECT" {
			return
		}
		p := ctx.Request.URL.Path
		if hasTrailingSlash(p) {
			p = p[:len(p)-1]
		} else {
			p = p + "/"
		}
		redirect(ctx, p, permanent)
	})
}

func NewRedirectCleanedPath(permanent bool) Handler {
	return HandlerFunc(func(ctx *Context) {
		if ctx.Routed() || ctx.Request.Method == "CONNECT" {
			return
		}
		p := ctx.Request.URL.Path
		if p == "" {
			p = "/"
		} else {
			if p[0] != '/' {
				p = "/" + p
			}
			p = path.Clean(p)
			if p != "/" && hasTrailingSlash(ctx.Request.URL.Path) {
				p += "/"
			}
		}
		if p != ctx.Request.URL.Path {
			redirect(ctx, p, permanent)
		}
	})
}

func NewAllowedMethods(handle bool) Handler {
	return HandlerFunc(func(ctx *Context) {
		if ctx.Routed() {
			return
		}
		methods := make([]string, 0, len(ctx.tree.methods))
		params := ctx.tree.getParams()
		for method, n := range ctx.tree.methods {
			if method == ctx.Request.Method || method == "" {
				continue
			}
			n, ok := n.match(ctx.Request.URL.Path, params[:0])
			if ok && n != nil && len(n.handlers) > 0 {
				methods = append(methods, method)
			}
		}
		ctx.tree.putParams(params)
		if len(methods) > 0 {
			sort.Strings(methods)
			ctx.Header().Set("Allow", strings.Join(methods, ", "))
			if handle {
				if ctx.Request.Method == "OPTIONS" {
					ctx.WriteHeader(http.StatusOK)
				} else {
					ctx.errorStatus(http.StatusMethodNotAllowed)
				}
			}
		} else if handle {
			ctx.NotFound()
		}
	})
}

func HandleNotFound(ctx *Context) {
	if !ctx.Routed() {
		ctx.NotFound()
	}
}
