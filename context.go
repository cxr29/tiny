// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/cxr29/log"
)

type Context struct {
	http.ResponseWriter
	Request       *http.Request
	Values        map[interface{}]interface{}
	Params        []string
	tree          *Tree
	node          *Node
	wroteHeader   bool
	written       int64
	status, index int
}

func (ctx *Context) Routed() bool {
	return ctx.node != nil
}

func (ctx *Context) call(i int) {
	var h Handler
	if j := i - len(ctx.tree.handlers); j < 0 {
		h = ctx.tree.handlers[i]
	} else if ctx.Routed() && j < len(ctx.node.handlers) {
		h = ctx.node.handlers[j]
	} else {
		return
	}
	h.ServeHTTP(ctx)
	if ctx.index == i && !ctx.wroteHeader {
		ctx.Next()
	}
}

func (ctx *Context) Next() {
	ctx.index++
	ctx.call(ctx.index)
}

func (ctx *Context) WriteHeader(code int) {
	if !ctx.wroteHeader {
		ctx.wroteHeader = true
		ctx.status = code
	}
	ctx.ResponseWriter.WriteHeader(code)
}

const contentType = "Content-Type"

func (ctx *Context) Write(data []byte) (int, error) {
	if !ctx.wroteHeader {
		h := ctx.Header()
		if h.Get("Transfer-Encoding") == "" && h.Get(contentType) == "" {
			h.Set(contentType, http.DetectContentType(data))
		}
		ctx.WriteHeader(http.StatusOK)
	}
	n, err := ctx.ResponseWriter.Write(data)
	log.ErrWarning(err)
	ctx.written += int64(n)
	return n, err
}

func (ctx *Context) WroteHeader() bool {
	return ctx.wroteHeader
}

func (ctx *Context) Written() int64 {
	return ctx.written
}

func (ctx *Context) Status() int {
	return ctx.status
}

func (ctx *Context) Param(name string) string {
	if ctx.Routed() {
		for k, v := range ctx.node.params {
			if v == name {
				return ctx.Params[k]
			}
		}
	}
	return ""
}

type valueKey struct {
	id uint
}

var (
	id uint
	mu sync.Mutex
)

func NewValueKey() (key valueKey) {
	mu.Lock()
	id++
	key.id = id
	mu.Unlock()
	return
}

func (ctx *Context) Value(k interface{}) interface{} {
	if v, ok := ctx.Values[k]; ok {
		return v
	} else {
		panic("key not exist")
	}
}

func (ctx *Context) SetValue(k, v interface{}) {
	if k == nil {
		panic("nil key")
	} else if !reflect.TypeOf(k).Comparable() {
		panic("not comparable key")
	} else if ctx.Values == nil {
		ctx.Values = map[interface{}]interface{}{k: v}
	} else if _, ok := ctx.Values[k]; ok {
		panic("key already exists")
	} else {
		ctx.Values[k] = v
	}
}
