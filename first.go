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

func (ctx *Context) FirstBool(k string) (b bool, n int) {
	k, n = ctx.First(k)
	if n > 0 {
		var err error
		b, err = strconv.ParseBool(k)
		if err != nil {
			n = -n
		}
	}
	return
}

func (ctx *Context) FirstInt(k string) (i int, n int) {
	k, n = ctx.First(k)
	if n > 0 {
		var err error
		i, err = strconv.Atoi(k)
		if err != nil {
			n = -n
		}
	}
	return
}

func (ctx *Context) FirstUint(k string) (u uint, n int) {
	k, n = ctx.First(k)
	if n > 0 {
		x, err := strconv.ParseUint(k, 10, 0)
		if err != nil {
			n = -n
		} else {
			u = uint(x)
		}
	}
	return
}

func (ctx *Context) FirstFloat64(k string) (f float64, n int) {
	k, n = ctx.First(k)
	if n > 0 {
		var err error
		f, err = strconv.ParseFloat(k, 64)
		if err != nil {
			n = -n
		}
	}
	return
}

func (ctx *Context) FirstFloat32(k string) (f float32, n int) {
	k, n = ctx.First(k)
	if n > 0 {
		x, err := strconv.ParseFloat(k, 32)
		if err != nil {
			n = -n
		} else {
			f = float32(x)
		}
	}
	return
}

func (ctx *Context) ParamBool(name string) (bool, error) {
	return strconv.ParseBool(ctx.Param(name))
}

func (ctx *Context) ParamInt(name string) (int, error) {
	return strconv.Atoi(ctx.Param(name))
}

func (ctx *Context) ParamFloat64(name string) (float64, error) {
	return strconv.ParseFloat(ctx.Param(name), 64)
}

func (ctx *Context) ParamFloat32(name string) (float32, error) {
	f, err := strconv.ParseFloat(ctx.Param(name), 32)
	return float32(f), err
}
