### route

[![GoDoc](http://godoc.org/github.com/frk/route?status.png)](http://godoc.org/github.com/frk/route)  [![Coverage](http://gocover.io/_badge/github.com/frk/route?nocache=route)](http://gocover.io/github.com/frk/route)


The package **route** provides an HTTP request multiplexer called **Router** that can be used as an alternative to Go's [http.ServeMux](http://golang.org/pkg/net/http/#ServeMux). This package is heavily inspired by [HttpRouter](https://github.com/julienschmidt/httprouter), [Gin Web Framework](https://github.com/gin-gonic/gin), and by Go's own [net/http](https://golang.org/pkg/net/http/) package.

**Requires Go 1.7+**

install with:

```sh
go get github.com/frk/route
```

documentation can be found at [GoDoc](http://godoc.org/github.com/frk/route).


## Overview

While this package is mostly analoguous to how http.ServeMux works, there is a small number of additional features that could be useful to some.

- register handlers for specific methods.
- specify parameter and catch-all segments in a handled pattern.
- specify dynamic parameters in the host of the handled pattern.


## Usage

**Basics** A GET specific HandleFunc.

```go
router.HandleFunc("GET", "/", func(c context.Context, w http.ResponseWriter, r *http.Request) {
	// ...
})
```
	
**Handle Multiple Methods** You can pass a list of methods separated by commas to
register the given handler to handle requests made with any one of those methods.

```go
router.HandleFunc("GET,POST,DELETE", "/users", func(c context.Context, w http.ResponseWriter, r *http.Request) {
	// ...
})
```

**Handle Any Method** You can use the "__\*__" as the method argument if you want
the given handler to handle requests made with *any* method.

```go
router.HandleFunc("*", "/posts", func(c context.Context, w http.ResponseWriter, r *http.Request) {
	// ...
})
```
	
**Handle Parameters (1)** You can specify *dynamic segments* in the pattern when
registering a handler. The router then, during handler resolution, extracts the
actual values from the request's URL path and stores them all in a `route.Params`
value inside the `context.Context`. Then the handler can access these values using
the `route.GetParams` function. Individual parameters are retrieved using the
"typed" methods of the `Params` type. Check the documentation for more info.

```go
router.HandleFunc("GET", "/posts/{post_slug}/comments/{comment_id}", func(c context.Context, w http.ResponseWriter, r *http.Request) {
	params := route.GetParams(c)
	slug, err := params.String("post_slug")
	if err != nil {
		// ...
	}
	comid := params.GetInt("comment_id")
	
	log.Printf("do something with post %s & comment %d\n", slug, comid)
})
```

**Handle Parameters (2)** You can specify a dynamic segment and a static segment
in the same part of a pattern, without conflict, as can be seen in this and the
previous example. The static segments take precedence over dynamic ones. For example
a GET request to `/posts/abc/comments/new` matches both, the pattern below, and the
one above, however since static segments take precedence the handler registered
below will be the one executed and the one above will be left alone.

```go
router.HandleFunc("GET", "/posts/{post_slug}/comments/new", func(c context.Context, w http.ResponseWriter, r *http.Request) {
	// ...
})
```
	
**Handle Catch-All Parameter** You can use "__\*__" to specify a *catch-all*
dynamic segment that matches different URL segments. The optional label after the
"__\*__" is used as the parameter's name and the parameter's value will be the part
of the URL that comes after the segment that's before the "__\*__", in the following
example the value of the `"filename"` parameter will contain everything that comes after
"/static/" e.g. "robots.txt.", "favicon.ico", as well as "assets/styles/app.css", etc.

```go
router.HandleFunc("GET", "/static/*filename", func(c context.Context, w http.ResponseWriter, r *http.Request) {
	params := route.GetParams(c)
	http.ServeFile(w, r, params.GetString("filename"))
})
```
		
**Custom 404 Handler** The method `SetNotFound` can be used to set the handler
that will be called every time a request's URL has no matching pattern registered
in the Router. By default the Router will use the `route.NotFound` HandlerFunc to
handle `404`s.

```go
router.SetNotFound(route.HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
	// ...
}))
```
