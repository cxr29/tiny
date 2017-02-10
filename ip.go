// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"net"
	"strings"
)

func (ctx *Context) ParseRemoteIP(realIP, forwardedFor bool) net.IP {
	if realIP {
		s := ctx.Request.Header.Get("X-Real-IP")
		ip := net.ParseIP(strings.TrimSpace(s))
		if ip != nil {
			return ip
		}
	}
	if forwardedFor {
		s := ctx.Request.Header.Get("X-Forwarded-For")
		if i := strings.Index(s, ","); i != -1 {
			s = s[:i]
		}
		ip := net.ParseIP(strings.TrimSpace(s))
		if ip != nil {
			return ip
		}
	}
	s, _, _ := net.SplitHostPort(ctx.Request.RemoteAddr)
	return net.ParseIP(s)
}

var keyIP = NewValueKey()

func (ctx *Context) SetRemoteIP(ip net.IP) {
	ctx.SetValue(keyIP, ip)
}

func (ctx *Context) RemoteIP() (ip net.IP) {
	if v, ok := ctx.Values[keyIP]; ok {
		ip = v.(net.IP)
	} else {
		ip = ctx.ParseRemoteIP(false, false)
		ctx.SetRemoteIP(ip)
	}
	return
}
