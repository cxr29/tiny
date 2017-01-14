Go tiny http router or web framework
===
Reinventing the wheel, reinventing the best wheel.

### Features
* lightweight and high performance
* work with tiny/http handlers
* support multiple/yield handlers
* take care for 501/trailing slash/cleaned path/405/404
* support any method/group subrouter
* named type/catch-all/regexp parameters
* context values/remote ip/first query/convenient methods/environment
* access log/compress

### Usage
```go
package main

import (
  "fmt"
  "net/http"

  "github.com/cxr29/log"
  "github.com/cxr29/tiny"
)

func main() {
  r := new(tiny.Router)

  // one or more handlers(handler is equivalent to middleware)
  r.Use(TinyHandler{}, HTTPHandler{})
  r.Use(TinyHandlerFunc, HTTPHandlerFunc)

  // call to yield after the rest handlers have been executed
  r.Use(func(ctx *tiny.Context) {
    fmt.Println("before")
    ctx.Next()
    fmt.Println("after")
  })

  // builtin handlers for no routes match
  //  501 Not Implemented
  //  try trailing slash, permanent redirect
  //  try cleaned path, permanent redirect
  //  405 Method Not Allowed
  //  404 Not Found
  r.Fallback()
  // or DIY
  r.Use(
    tiny.HandleNotImplemented,
    tiny.NewRedirectTrailingSlash(false), // try trailing slash, temporary redirect
    tiny.NewRedirectCleanedPath(false),   // try cleaned path, temporary redirect
    tiny.NewAllowedMethods(false),        // no auto-handle, just set the Allow header
    func(ctx *tiny.Context) {
      // already routed
      if ctx.Routed() {
        return
      }

      allowedMethods := ctx.Header().Get("Allow")
      if allowedMethods == "" { // no allowed methods
        ctx.NotFound()
        return
      }

      // comma-separated allowed methods
      fmt.Println(allowedMethods)

      if ctx.Request.Method == "OPTIONS" { // 200
        ctx.WriteHeader(http.StatusOK)
      } else { // 405
        ctx.WriteHeader(http.StatusMethodNotAllowed)
        ctx.WriteString(http.StatusText(http.StatusMethodNotAllowed))
      }
    },
  )

  // method handlers
  r.GET("/", handlers...)
  r.POST("/foo", handlers...)
  // ...

  // any method handlers
  r.Any("/foo", handlers...) // only when no explicit method routes match

  // group handlers
  r.Group("/foo/bar", func(r *tiny.Router) {
    // subrouter just the same
  }, handlers...) // handlers called before the subrouter's handlers

  // named parameters match anything except slashes
  //  /foo      match
  //  /foo/     no match
  //  /foo/bar  no match
  r.GET("/<name>", func(ctx *tiny.Context) {
    // get param by name
    fmt.Println(ctx.Param("name"))
    // or by index
    fmt.Println(ctx.Params[0])
  })

  // named type parameters
  // boolean match true/false
  r.GET("/<long:boolean>/<short:bool>/<even:b>")

  // integer match well-formed integers
  r.GET("/<long:integer>/<short:int>/<even:i>")

  // number match well-formed numbers
  r.GET("/<long:number>/<short:num>/<even:n>")

  // string/catch-all match anything include slashes
  r.GET("/<long:string>/<short:str>/<even:s>")

  // caret represents begin match the regexp
  r.GET(`/<name^[0-9]+>`)

  log.ErrFatal(tiny.ListenAndServe(":8080", r.Handler()))
}
```

##### I hate writing documentation but [RTFSC](https://godoc.org/github.com/cxr29/tiny).
##### I hate writing test cases but I have tested it. I did my best.
