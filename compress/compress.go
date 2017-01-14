// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/cxr29/log"
	"github.com/cxr29/tiny"
)

const (
	acceptEncoding   = "Accept-Encoding"
	acceptRanges     = "Accept-Ranges"
	contentEncoding  = "Content-Encoding"
	contentLength    = "Content-Length"
	contentRange     = "Content-Range"
	contentType      = "Content-Type"
	transferEncoding = "Transfer-Encoding"
	vary             = "Vary"
)

type Options struct {
	Level   int
	Gzip    bool
	Deflate bool
}

var (
	DefaultOptions = &Options{flate.BestSpeed, true, true}
	keyCompress    = tiny.NewValueKey()
)

func Pull(ctx *tiny.Context) bool {
	return ctx.Value(keyCompress).(bool)
}

func On(ctx *tiny.Context) {
	ctx.Values[keyCompress] = true
}

func Off(ctx *tiny.Context) {
	ctx.Values[keyCompress] = false
}

func New(o *Options) tiny.HandlerFunc {
	if o == nil {
		o = DefaultOptions
	}
	return func(ctx *tiny.Context) {
		s := strings.ToLower(ctx.Request.Header.Get(acceptEncoding))
		if o.Gzip && strings.Contains(s, "gzip") {
			s = "gzip"
		} else if o.Deflate && strings.Contains(s, "deflate") {
			s = "deflate"
		} else {
			return
		}
		rw := &responseWriter{
			ResponseWriter: ctx.ResponseWriter,
			ctx:            ctx,
			o:              o,
			ce:             s,
		}
		ctx.ResponseWriter = rw
		On(ctx)
		defer func() {
			if rw.flag == 1 {
				log.ErrError(rw.wc.Close())
			}
		}()
		ctx.Next()
	}
}

type responseWriter struct {
	http.ResponseWriter
	ctx  *tiny.Context
	o    *Options
	ce   string
	flag int8
	wc   io.WriteCloser
}

func allowedStatusCode(code int) bool {
	switch {
	case code >= 100 && code <= 199:
		return false
	case code == 204 || code == 304:
		return false
	}
	return true
}

func allowedContentType(c string) bool {
	var t string
	if i := strings.Index(c, "/"); i >= 0 {
		t = c[i+1:]
		c = c[:i]
	}
	for _, i := range [...]string{"image", "audio", "video"} {
		if c == i {
			return false
		}
	}
	if c == "application" {
		for _, i := range [...]string{"ogg", "x-rar-compressed", "zip", "x-gzip"} {
			if t == i {
				return false
			}
		}
	}
	return true
}

func (rw *responseWriter) newWriter() bool {
	var err error
	switch rw.ce {
	case "gzip":
		rw.wc, err = gzip.NewWriterLevel(rw.ResponseWriter, rw.o.Level)
	case "deflate":
		rw.wc, err = flate.NewWriter(rw.ResponseWriter, rw.o.Level)
	}
	if err == nil {
		return true
	}
	log.Errorln(err)
	return false
}

func (rw *responseWriter) initialize(data []byte) {
	if rw.flag != 0 {
		return
	}
	if Pull(rw.ctx) {
		h := rw.Header()
		if h.Get(contentEncoding) == "" && h.Get(contentRange) == "" {
			ct := h.Get(contentType)
			if h.Get(transferEncoding) == "" && ct == "" {
				ct = http.DetectContentType(data)
				h.Set(contentType, ct)
			}
			if allowedContentType(ct) && rw.newWriter() {
				h.Del(contentLength)
				h.Del(acceptRanges)
				h.Set(contentEncoding, rw.ce)
				h.Set(vary, acceptEncoding)
				rw.flag = 1
				return
			}
		}
	}
	rw.flag = -1
}

func (rw *responseWriter) WriteHeader(code int) {
	if allowedStatusCode(code) {
		rw.initialize(nil)
	} else if rw.flag == 0 {
		rw.flag = -1
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.initialize(data)
	if rw.flag == 1 {
		return rw.wc.Write(data)
	} else {
		return rw.ResponseWriter.Write(data)
	}
}
