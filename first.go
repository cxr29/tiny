// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"strconv"

	"github.com/cxr29/log"
)

func (ctx *Context) First(k string) (s string, n int) {
	log.ErrWarning(ctx.Request.ParseForm())
	a := ctx.Request.Form[k]
	if n = len(a); n > 0 {
		s = a[0]
	}
	return
}

func (ctx *Context) FirstBool(k string) (bool, int) {
	k, n := ctx.First(k)
	if n > 0 {
		b, err := strconv.ParseBool(k)
		if err == nil {
			return b, n
		}
	}
	return false, -n
}

func (ctx *Context) FirstInt(k string) (int, int) {
	k, n := ctx.First(k)
	if n > 0 {
		i, err := strconv.Atoi(k)
		if err == nil {
			return i, n
		}
	}
	return 0, -n
}

func (ctx *Context) FirstUint(k string) (uint, int) {
	k, n := ctx.First(k)
	if n > 0 {
		u, err := strconv.ParseUint(k, 10, 0)
		if err == nil {
			return uint(u), n
		}
	}
	return 0, -n
}

func (ctx *Context) FirstFloat64(k string) (float64, int) {
	k, n := ctx.First(k)
	if n > 0 {
		f, err := strconv.ParseFloat(k, 64)
		if err == nil {
			return f, n
		}
	}
	return 0, -n
}

func (ctx *Context) FirstFloat32(k string) (float32, int) {
	k, n := ctx.First(k)
	if n > 0 {
		f, err := strconv.ParseFloat(k, 32)
		if err == nil {
			return float32(f), n
		}
	}
	return 0, -n
}

func (ctx *Context) ParamBool(name string) (bool, bool) {
	b, err := strconv.ParseBool(ctx.Param(name))
	if err != nil {
		return false, false
	}
	return b, true
}

func (ctx *Context) ParamInt(name string) (int, bool) {
	i, err := strconv.Atoi(ctx.Param(name))
	if err != nil {
		return 0, false
	}
	return i, true
}

func (ctx *Context) ParamUint(name string) (uint, bool) {
	u, err := strconv.ParseUint(ctx.Param(name), 10, 0)
	if err != nil {
		return 0, false
	}
	return uint(u), true
}

func (ctx *Context) ParamFloat64(name string) (float64, bool) {
	f, err := strconv.ParseFloat(ctx.Param(name), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (ctx *Context) ParamFloat32(name string) (float32, bool) {
	f, err := strconv.ParseFloat(ctx.Param(name), 32)
	if err != nil {
		return 0, false
	}
	return float32(f), true
}
