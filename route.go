// The package route provides an HTTP request multiplexer called Router that
// can be used as an alternative to Go's http.ServeMux.
package route

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"
)

// Router is an HTTP request router. It matches the URL of each incoming
// request against a list of registered patterns and calls the handler for the
// pattern that most closely matches the URL.
type Router struct {
	mu    sync.RWMutex
	hosts bool
	root  *node

	handle404 Handler

	ctxpool sync.Pool
}

// NewRouter allocates and returns a new Router.
func NewRouter() *Router {
	r := &Router{}
	r.root = &node{}
	r.handle404 = HandlerFunc(NotFound)

	r.ctxpool.New = func() interface{} {
		return &ctx{Params{}}
	}
	return r
}

// ServeHTTP dispatches the request to the handler whose pattern most closely
// matches the request URL. ServeHTTP implements the http.Handler interface.
//
// ServeHTTP also instantiates a request-scoped context.Context that holds the
// request specific Params value which can be retrieved using the GetParams function.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "*" {
		if req.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		c        = r.ctxpool.Get().(*ctx)
		po       = c.Params
		h, ps, _ = r.handler(req, po)
	)

	c.Params = ps
	h.ServeHTTP(c, w, req)

	r.ctxpool.Put(c)
}

// Handler returns the Handler and Params to use for the given request, consulting
// r.URL.Path and r.Method. If there is no Handler registered for the request's
// path and method a not-found Handler will be returned. Handler is guaranteed to
// always return a non-nil handler.
//
// Handler also returns the registered pattern that matches the request.
func (r *Router) Handler(req *http.Request) (h Handler, ps Params, pat string) {
	return r.handler(req, Params{})
}

type tsr int

const (
	tsrNone tsr = iota
	tsrWithSlash
	tsrWithoutSlash
)

func (r *Router) handler(req *http.Request, po Params) (h Handler, ps Params, pat string) {
	var (
		host  = req.Host
		path  = req.URL.Path
		redir tsr
	)
	if r.hosts {
		h, ps, pat, redir = r.root.lookup(host+path, po)
	}
	if h == nil && redir == tsrNone {
		h, ps, pat, redir = r.root.lookup(path, po)
	}
	if h == nil {
		if redir == tsrWithSlash {
			h = RedirectHandler(path+"/", http.StatusMovedPermanently)
		} else if redir == tsrWithoutSlash {
			h = RedirectHandler(path[:len(path)-1], http.StatusMovedPermanently)
		} else {
			h = r.handle404
		}
	}

	return h, ps, pat
}

// Handle registers the handler for the given pattern and method. If a handler
// already exists for that pattern and method, Handle panics.
func (r *Router) Handle(method, pattern string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pattern == "" {
		panic("route.Handle: empty pattern")
	}
	if method == "" {
		panic("route.Handle: empty method")
	}
	if handler == nil {
		panic("route.Handle: nil handler")
	}
	if pattern[0] != '/' {
		r.hosts = true
	}

	if err := r.root.insert(method, pattern, handler); err != nil {
		panic(fmt.Sprintf("route.Handle: %s %s: %v", method, pattern, err))
	}
}

// HandleFunc registers the handler function for the given pattern and method.
func (r *Router) HandleFunc(method, pattern string, handler func(context.Context, http.ResponseWriter, *http.Request)) {
	r.Handle(method, pattern, HandlerFunc(handler))
}

// SetNotFound installs the Router's NotFound handler to be used when there is no
// pattern registered that matches a reqeust's URL path.
func (r *Router) SetNotFound(h Handler) {
	if h != nil {
		r.handle404 = h
	}
}

// Handler is analoguous to go's standard net/http.Handler
//
// Objects implementing the Handler interface can be registered to serve a
// particular path and method in the HTTP server.
type Handler interface {
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

// HandlerFunc is analoguous to go's standard net/http.HandlerFunc
//
// The HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

// ServeHTTP calls f(w, r, p).
func (f HandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// NotFound replies to the request with an HTTP 404 not found error. If the request
// path is not in its canonical form the request will be redirected to the canonical path.
func NotFound(_ context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		clean := cleanPath(r.URL.Path)
		if clean != r.URL.Path && clean != r.Referer() {
			http.Redirect(w, r, clean, http.StatusMovedPermanently)
			return
		}
	}
	http.NotFound(w, r)
}

// Redirect to a fixed URL
type redirectHandler struct {
	url  string
	code int
}

func (rh *redirectHandler) ServeHTTP(_ context.Context, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, rh.url, rh.code)
}

// Copied from Go's net/http.RedirectHandler
//
// RedirectHandler returns a request handler that redirects
// each request it receives to the given url using the given
// status code.
//
// The provided code should be in the 3xx range and is usually
// StatusMovedPermanently, StatusFound or StatusSeeOther.
func RedirectHandler(url string, code int) Handler {
	return &redirectHandler{url, code}
}

// ctxKey is an unexported package specific type for context.Context value keys
// to prevent collisions with keys defined in other packages.
type ctxKey int

// paramsKey is the key for route.Params values in Contexts. Clients should use
// route.Context and route.GetParams instead of using this key directly.
const paramsKey ctxKey = 0

// Context returns a copy of parent which carries the Params value p.
func Context(parent context.Context, p Params) context.Context {
	return context.WithValue(parent, paramsKey, p)
}

// GetParams returns the Params value stored in ctx. If no Params value is stored
// in the ctx GetParams returns an empty, non-nil Params value.
func GetParams(c context.Context) Params {
	if c != nil {
		if p, ok := c.Value(paramsKey).(Params); ok {
			return p
		}
	}
	return Params{}
}

// cleanPath is copied from net/http/server.go.
// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root,
	// put trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// The ctx type implements the context.Context interface.
type ctx struct {
	Params Params
}

func (c *ctx) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c *ctx) Done() <-chan struct{} {
	return nil
}

func (c *ctx) Err() error {
	return nil
}

func (c *ctx) Value(key interface{}) interface{} {
	return c.Params
}