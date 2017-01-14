// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/cxr29/log"
)

func (ctx *Context) WriteString(s string) (int, error) {
	ctx.ContentTypePlain()
	return ctx.Write([]byte(s))
}

func (ctx *Context) WriteJSON(v interface{}) (int, error) {
	ctx.ContentTypeJSON()
	p, err := json.Marshal(v)
	if err != nil {
		log.Warningln(err)
		return 0, err
	}
	return ctx.Write(p)
}

func (ctx *Context) WriteXML(v interface{}) (int, error) {
	ctx.ContentTypeXML()
	p, err := xml.Marshal(v)
	if err != nil {
		log.Warningln(err)
		return 0, err
	}
	return ctx.Write(p)
}

func (ctx *Context) WriteData(v interface{}) (int, error) {
	return ctx.WriteJSON(map[string]interface{}{"Data": v})
}

func (ctx *Context) IsAJAX() bool {
	return ctx.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (ctx *Context) writeError(s string) (int, error) {
	if ctx.IsAJAX() {
		return ctx.WriteJSON(map[string]string{"Error": s})
	}
	return ctx.WriteString("Error: " + s)
}

func (ctx *Context) WriteError(e interface{}) (int, error) {
	return ctx.writeError(fmt.Sprint(e))
}

func (ctx *Context) WriteErrorf(format string, a ...interface{}) (int, error) {
	return ctx.writeError(fmt.Sprintf(format, a...))
}

func (ctx *Context) DecodeJSON(v interface{}) error {
	defer ctx.Request.Body.Close()
	return json.NewDecoder(ctx.Request.Body).Decode(v)
}

func (ctx *Context) DecodeXML(v interface{}) error {
	defer ctx.Request.Body.Close()
	return xml.NewDecoder(ctx.Request.Body).Decode(v)
}

func (ctx *Context) ServeFile(name string) {
	http.ServeFile(ctx, ctx.Request, name)
}

func (ctx *Context) ContentLength(i int) {
	ctx.Header().Set("Content-Length", strconv.Itoa(i))
}

func (ctx *Context) ContentType(s string) {
	ctx.Header().Set(contentType, s)
}

func (ctx *Context) utf8ContentType(s string) {
	ctx.ContentType(s + "; charset=utf-8")
}

func (ctx *Context) ContentTypePlain() {
	ctx.utf8ContentType("text/plain")
}

func (ctx *Context) ContentTypeHTML() {
	ctx.utf8ContentType("text/html")
}

func (ctx *Context) ContentTypeJSON() {
	ctx.utf8ContentType("application/json")
}

func (ctx *Context) ContentTypeXML() {
	ctx.utf8ContentType("application/xml")
}

func (ctx *Context) ContentTypeCSV() {
	ctx.utf8ContentType("text/csv")
}

func (ctx *Context) ContentTypeXLSX() {
	ctx.ContentType("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
}

const filenameLength = 255

var filenameRegexp = regexp.MustCompile(`^[-.0-9A-Z_a-z]+$`)

func (ctx *Context) ContentDisposition(filename, fallback string) {
	if len(filename) == 0 || len(filename) > filenameLength {
		panic("malformed filename")
	} else if filenameRegexp.MatchString(filename) {
		filename = "attachment; filename=" + filename
	} else {
		filename = "attachment; filename*=UTF-8''" + url.QueryEscape(filename)
		if len(fallback) > 0 {
			if len(fallback) > filenameLength || !filenameRegexp.MatchString(fallback) {
				panic("malformed fallback")
			}
			filename += "; filename=" + fallback
		}
	}
	ctx.Header().Set("Content-Disposition", filename)
}

func (ctx *Context) MaxAge(seconds int) {
	v := time.Now().Add(time.Duration(seconds) * time.Second).Format(http.TimeFormat)
	h := ctx.Header()
	h.Set("Expires", v)
	h.Set("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
}

func (ctx *Context) NoCache() {
	h := ctx.Header()
	h.Set("Expires", "0")
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Pragma", "no-cache")
}

func (ctx *Context) LastModified(t time.Time) {
	ctx.Header().Set("Last-Modified", t.Format(http.TimeFormat))
}

func (ctx *Context) IfModifiedSince(t time.Time) bool {
	v, _ := http.ParseTime(ctx.Request.Header.Get("If-Modified-Since"))
	return v.IsZero() || v.Unix() != t.Unix()
}

func (ctx *Context) ETag(s string) {
	ctx.Header().Set("ETag", s)
}

func (ctx *Context) IfNoneMatch(s string) bool {
	v := ctx.Request.Header.Get("If-None-Match")
	return v == "" || v != s
}

func (ctx *Context) MovedPermanently(location string) {
	http.Redirect(ctx, ctx.Request, location, http.StatusMovedPermanently)
}

func (ctx *Context) Found(location string) {
	http.Redirect(ctx, ctx.Request, location, http.StatusFound)
}

func (ctx *Context) NotModified() {
	ctx.WriteHeader(http.StatusNotModified)
}

func (ctx *Context) errorStatus(code int) {
	http.Error(ctx, http.StatusText(code), code)
}

func (ctx *Context) BadRequest() {
	ctx.errorStatus(http.StatusBadRequest)
}

func (ctx *Context) Forbidden() {
	ctx.errorStatus(http.StatusForbidden)
}

func (ctx *Context) NotFound() {
	ctx.errorStatus(http.StatusNotFound)
}

func (ctx *Context) InternalServerError() {
	ctx.errorStatus(http.StatusInternalServerError)
}

func (ctx *Context) ServiceUnavailable() {
	ctx.errorStatus(http.StatusServiceUnavailable)
}
