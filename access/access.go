// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package access

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cxr29/log"
	"github.com/cxr29/tiny"
)

type Options struct {
	Before, After  io.Writer
	Format, Layout string
	Comma          rune
	UseCRLF        bool
	Fields         []string
	mu             sync.Mutex
	n              int
}

func (o *Options) add(i int) (n int) {
	o.mu.Lock()
	o.n += i
	n = o.n
	o.mu.Unlock()
	return
}

var DefaultOptions = &Options{
	After:  os.Stdout,
	Format: "text",
	Layout: "2006-01-02 15:04:05",
	Comma:  ',',
	Fields: []string{
		"time",
		"addr",
		"method",
		"host",
		"uri",
		"proto",
		"referer",
		"ua",
		"panic",
		"status",
		"size",
		"duration",
		"count",
		"r_x_forwarded_for",
	},
}

var keyAccess = tiny.NewValueKey()

func Pull(ctx *tiny.Context) *Access {
	return ctx.Value(keyAccess).(*Access)
}

func underscore2hyphen(s string) string {
	return http.CanonicalHeaderKey(strings.Replace(s[2:], "_", "-", -1))
}

func New(o *Options) tiny.HandlerFunc {
	if o == nil {
		o = DefaultOptions
	}
	switch o.Format {
	case "text", "csv", "json":
	default:
		panic("unsupported format")
	}
	var cookies []string
	var reqHeaders, resHeaders map[string]string
	for _, s := range o.Fields {
		if strings.HasPrefix(s, "c_") {
			cookies = append(cookies, s)
		} else if strings.HasPrefix(s, "r_") {
			if reqHeaders == nil {
				reqHeaders = map[string]string{s: underscore2hyphen(s)}
			} else {
				reqHeaders[s] = underscore2hyphen(s)
			}
		} else if strings.HasPrefix(s, "w_") {
			if resHeaders == nil {
				resHeaders = map[string]string{s: underscore2hyphen(s)}
			} else {
				resHeaders[s] = underscore2hyphen(s)
			}
		}
	}
	return func(ctx *tiny.Context) {
		a := &Access{o: o, m: make(map[string]interface{}, len(o.Fields))}
		ctx.SetValue(keyAccess, a)
		if o.Before == nil && o.After == nil {
			return
		}
		t := time.Now()
		a.Set("count", o.add(1))
		a.Set("time", t.Format(o.Layout))
		a.Set("addr", ctx.RemoteIP().String())
		a.Set("method", ctx.Request.Method)
		a.Set("host", ctx.Request.Host)
		a.Set("uri", ctx.Request.URL.RequestURI())
		a.Set("proto", ctx.Request.Proto)
		a.Set("referer", ctx.Request.Referer())
		a.Set("ua", ctx.Request.UserAgent())
		for _, i := range cookies {
			if c, err := ctx.Request.Cookie(i[2:]); err == nil {
				a.Set(i, c.Value)
			}
		}
		a.setHeaders(ctx.Request.Header, reqHeaders)
		if o.Before != nil {
			a.writeTo(o.Before)
		}
		if o.After == nil {
			return
		}
		defer func() {
			if a.Off {
				return
			}
			err := recover()
			a.Set("panic", err != nil)
			a.Set("status", ctx.Status())
			a.Set("size", ctx.Written())
			a.Set("duration", time.Since(t))
			a.setHeaders(ctx.Header(), resHeaders)
			a.Set("count", o.add(-1))
			if o.After != nil {
				a.writeTo(o.After)
			}
			if err != nil {
				panic(err)
			}
		}()
		ctx.Next()
	}
}

type Access struct {
	Off bool
	o   *Options
	m   map[string]interface{}
}

func (a *Access) Get(k string) interface{} {
	return a.m[k]
}

func (a *Access) Set(k string, v interface{}) {
	a.m[k] = v
}

func (a *Access) Del(k string) {
	delete(a.m, k)
}

func (a *Access) setHeaders(h http.Header, m map[string]string) {
	for k, v := range m {
		if b, ok := h[v]; ok {
			a.Set(k, strings.Join(b, ", "))
		}
	}
}

func (a *Access) writeTo(w io.Writer) {
	var err error
	switch a.o.Format {
	case "text":
		_, err = w.Write(a.text())
	case "csv":
		c := csv.NewWriter(w)
		c.Comma = a.o.Comma
		c.UseCRLF = a.o.UseCRLF
		if err = c.Write(a.csv()); err == nil {
			c.Flush()
			err = c.Error()
		}
	case "json":
		_, err = w.Write(a.json())
	}
	log.ErrError(err)
}

func (a *Access) text() []byte {
	var b bytes.Buffer
	for i, s := range a.o.Fields {
		if i > 0 {
			b.WriteByte('\t')
		}
		if v := a.m[s]; v != nil {
			s = fmt.Sprint(v)
			if strings.ContainsAny(s, "\x00\t\n\v\f\r\\") {
				for i, n := 0, len(s); i < n; i++ {
					switch s[i] {
					case '\x00':
						b.WriteString(`\0`)
					case '\t':
						b.WriteString(`\t`)
					case '\n':
						b.WriteString(`\n`)
					case '\v':
						b.WriteString(`\v`)
					case '\f':
						b.WriteString(`\f`)
					case '\r':
						b.WriteString(`\r`)
					case '\\':
						b.WriteString(`\\`)
					default:
						b.WriteByte(s[i])
					}
				}
			} else {
				b.WriteString(s)
			}
		}
	}
	b.WriteByte('\n')
	return b.Bytes()
}

func (a *Access) csv() []string {
	b := make([]string, len(a.o.Fields))
	for i, s := range a.o.Fields {
		if v := a.m[s]; v != nil {
			b[i] = fmt.Sprint(v)
		}
	}
	return b
}

func (a *Access) json() []byte {
	p, err := json.Marshal(a.m)
	log.ErrError(err)
	return append(p, '\n')
}
