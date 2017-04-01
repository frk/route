package route

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type strHandler string

func (h strHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Handled-By", string(h))
	recordContext(c, w)
}

type routerSetup []struct {
	method  string
	pattern string
	handler strHandler
}

func (setup routerSetup) Router() *Router {
	r := NewRouter()
	for _, t := range setup {
		r.Handle(t.method, t.pattern, t.handler)
	}
	return r
}

type routerTests []struct {
	method  string
	path    string
	handler string
	code    int
	params  Params
	pattern string
}

func (tests routerTests) Run(t *testing.T, router *Router) {
	for i, tt := range tests {
		r := mustNewRequest(tt.method, tt.path, nil)

		w := newRecorder()
		router.ServeHTTP(w, r)
		equals(t, i, w.HeaderMap.Get("Handled-By"), tt.handler)
		equals(t, i, w.Code, tt.code)
		equals(t, i, w.Params(), tt.params)

		_, _, pat := router.Handler(r)
		equals(t, i, pat, tt.pattern)
	}
}

func equals(t *testing.T, i int, got, want interface{}) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("#%d: got %v, want %v", i, got, want)
	}
}

func mustNewRequest(method, urlStr string, body io.Reader) *http.Request {
	r, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(err)
	}
	return r
}

type ctxRecorder struct {
	*httptest.ResponseRecorder
	Ctx context.Context
}

func newRecorder() *ctxRecorder {
	return &ctxRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (cr *ctxRecorder) Params() Params {
	return GetParams(cr.Ctx)
}

func recordContext(ctx context.Context, w http.ResponseWriter) {
	if rec, ok := w.(*ctxRecorder); ok {
		rec.Ctx = ctx
	}
}

func TestRouterHandle_PanicsWithInvalidArgs(t *testing.T) {
	//t.Skip()
	var router = routerSetup{
		{"GET", "/foo", "tt"},
		{"GET", "/foo/{bar_id}", "tt"},
	}.Router()

	var tests = []struct {
		method    string
		pattern   string
		handler   Handler
		wantPanic string
	}{
		{
			method:    "GET",
			pattern:   "",
			handler:   strHandler("test"),
			wantPanic: "route.Handle: empty pattern",
		}, {
			method:    "GET",
			pattern:   "/foo",
			handler:   strHandler("test"),
			wantPanic: "route.Handle: GET /foo: " + (&routeError{typ: errMethodConflict, a: "GET"}).Error(),
		}, {
			method:    "GET",
			pattern:   "/foo/{bar_name}",
			handler:   strHandler("test"),
			wantPanic: "route.Handle: GET /foo/{bar_name}: " + (&routeError{errParamConflict, "bar_name", "bar_id"}).Error(),
		}, {
			method:    "GET",
			pattern:   "/foo/bar",
			handler:   nil,
			wantPanic: "route.Handle: nil handler",
		}, {
			method:    "",
			pattern:   "/foo/bar",
			handler:   strHandler("test"),
			wantPanic: "route.Handle: empty method",
		}, {
			method:    "GET,POST,,PUT",
			pattern:   "/foo/bar",
			handler:   strHandler("test"),
			wantPanic: "route.Handle: GET,POST,,PUT /foo/bar: Missing method",
		},
	}

	for _, tt := range tests {
		func() {
			defer func() {
				if got := recover(); got != tt.wantPanic {
					t.Errorf("got %v, want %q", got, tt.wantPanic)
				}
			}()
			router.Handle(tt.method, tt.pattern, tt.handler)
		}()
	}
}

func TestRouterServeHTTP_Static(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/", "handler_a"},
		{"GET", "/foo", "handler_b"},
		{"GET", "/foo/bar", "handler_c"},
		{"GET", "/foo/bar/baz", "handler_d"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/",
			handler: "handler_a", code: 200,
			params: Params{}, pattern: "/",
		}, {
			method: "GET", path: "/foo",
			handler: "handler_b", code: 200,
			params: Params{}, pattern: "/foo",
		}, {
			method: "GET", path: "/foo/bar",
			handler: "handler_c", code: 200,
			params: Params{}, pattern: "/foo/bar",
		}, {
			method: "GET", path: "/foo/bar/baz",
			handler: "handler_d", code: 200,
			params: Params{}, pattern: "/foo/bar/baz",
		},
	}.Run(t, router)
}

func TestRouterServeHTTP_TSR(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar", "handler_a"},
		{"GET", "/foo/bar/a", "handler_b"},
		{"GET", "/foo/bar/{b}", "handler_c"},
		{"GET", "/foo/{c}/", "handler_d"},
		{"GET", "/foo/bar/baz", "handler_e"},

		{"GET", "/foo/baz/", "handler_f"},
		{"GET", "/foo/bazz", "handler_g"},

		{"GET", "/aaa/{a}", "h"},
		{"GET", "/bbb/{b}/", "h"},
		{"GET", "/ccc/foo", "h"},
		{"GET", "/ddd/foo/", "h"},
	}.Router()

	tests := []struct {
		method  string
		path    string
		handler Handler
	}{
		{
			method: "GET", path: "/foo/bar/",
			handler: RedirectHandler("/foo/bar", 301),
		}, {
			method: "GET", path: "/foo/bar/baz/",
			handler: RedirectHandler("/foo/bar/baz", 301),
		}, {
			method: "GET", path: "/foo/baz",
			handler: RedirectHandler("/foo/baz/", 301),
		}, {
			method: "GET", path: "/aaa/foo/",
			handler: RedirectHandler("/aaa/foo", 301),
		}, {
			method: "GET", path: "/bbb/foo",
			handler: RedirectHandler("/bbb/foo/", 301),
		}, {
			method: "GET", path: "/ccc/foo/",
			handler: RedirectHandler("/ccc/foo", 301),
		}, {
			method: "GET", path: "/ddd/foo",
			handler: RedirectHandler("/ddd/foo/", 301),
		},
	}

	for i, tt := range tests {
		r := mustNewRequest(tt.method, tt.path, nil)
		h, _, _ := router.Handler(r)

		equals(t, i, h, tt.handler)
	}
}

func TestRouterServeHTTP_Param(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar/baz", "handler_a"},
		{"GET", "/foo/{b}/baz", "hanlder_b"},
		{"GET", "/foo/bar/{c}", "handler_c"},
		{"GET", "/foo/{b}/{c}", "handler_d"},
		{"GET", "/{a}/{b}/{c}", "handler_e"},
		{"GET", "/{a}/{b}/baz", "handler_f"},
		{"GET", "/{a}/bar/baz", "handler_g"},
		{"GET", "/{a}/bar/{c}", "handler_h"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar/baz",
			handler: "handler_a", code: 200,
			params: Params{}, pattern: "/foo/bar/baz",
		}, {
			method: "GET", path: "/foo/y/baz",
			handler: "hanlder_b", code: 200,
			params: Params{{"b", "y"}}, pattern: "/foo/{b}/baz",
		}, {
			method: "GET", path: "/foo/bar/z",
			handler: "handler_c", code: 200,
			params: Params{{"c", "z"}}, pattern: "/foo/bar/{c}",
		}, {
			method: "GET", path: "/foo/y/z",
			handler: "handler_d", code: 200,
			params: Params{{"b", "y"}, {"c", "z"}}, pattern: "/foo/{b}/{c}",
		}, {
			method: "GET", path: "/x/y/z",
			handler: "handler_e", code: 200,
			params: Params{{"a", "x"}, {"b", "y"}, {"c", "z"}}, pattern: "/{a}/{b}/{c}",
		}, {
			method: "GET", path: "/x/y/baz",
			handler: "handler_f", code: 200,
			params: Params{{"a", "x"}, {"b", "y"}}, pattern: "/{a}/{b}/baz",
		}, {
			method: "GET", path: "/x/bar/baz",
			handler: "handler_g", code: 200,
			params: Params{{"a", "x"}}, pattern: "/{a}/bar/baz",
		}, {
			method: "GET", path: "/x/bar/z",
			handler: "handler_h", code: 200,
			params: Params{{"a", "x"}, {"c", "z"}}, pattern: "/{a}/bar/{c}",
		}, {
			// NOTE(mkopriva): this case, as opposed to the previous one,
			// checks that 'b' in the third segment matches the {c} param node
			// instead of the 'baz' static node which would happend if
			// we matched using only the node.indices in lookup.
			method: "GET", path: "/x/bar/b",
			handler: "handler_h", code: 200,
			params: Params{{"a", "x"}, {"c", "b"}}, pattern: "/{a}/bar/{c}",
		},
	}.Run(t, router)
}

func TestRouterServeHTTP_MixedStaticParamsSegments(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar", "handler_a"},
		{"GET", "/fou/bar", "handler_b"},
		{"GET", "/fou/bus", "handler_c"},
		{"GET", "/{x}", "handler_d"},
		{"GET", "/fou/{x}", "handler_e"},
		{"GET", "/{x}/bar", "handler_f"},
		{"GET", "/{x}/bat", "handler_g"},
		{"GET", "/{x}/{y}", "handler_h"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar",
			handler: "handler_a", code: 200,
			params: Params{}, pattern: "/foo/bar",
		}, {
			method: "GET", path: "/fou/bar",
			handler: "handler_b", code: 200,
			params: Params{}, pattern: "/fou/bar",
		}, {
			method: "GET", path: "/fou/bus",
			handler: "handler_c", code: 200,
			params: Params{}, pattern: "/fou/bus",
		}, {
			method: "GET", path: "/abc",
			handler: "handler_d", code: 200,
			params: Params{{"x", "abc"}}, pattern: "/{x}",
		}, {
			method: "GET", path: "/fox",
			handler: "handler_d", code: 200,
			params: Params{{"x", "fox"}}, pattern: "/{x}",
		}, {
			method: "GET", path: "/fou/bat",
			handler: "handler_e", code: 200,
			params: Params{{"x", "bat"}}, pattern: "/fou/{x}",
		}, {
			method: "GET", path: "/fox/bag",
			handler: "handler_h", code: 200,
			params: Params{{"x", "fox"}, {"y", "bag"}}, pattern: "/{x}/{y}",
		},
	}.Run(t, router)
}

func TestRouterServeHTTP_CatchAll(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar/baz", "handler_a"},
		{"GET", "/foo/bar/*abc", "handler_b"},
		{"GET", "/foo/*abc", "handler_c"},
		{"GET", "/*abc", "handler_d"},
		{"GET", "/goo/car/*", "handler_e"}, // catch-all "name" is optional

		// NOTE(mkopriva): param takes precedence over catch-all and
		// therefore the /goo/*abc case will never be "hit". Might be
		// a good idea to panic in this scenario.
		{"GET", "/goo/{b}", "handler_f"},
		{"GET", "/goo/*abc", "handler_g"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar/baz",
			handler: "handler_a", code: 200,
			params: Params{}, pattern: "/foo/bar/baz",
		}, {
			method: "GET", path: "/foo/bar/x/y/z",
			handler: "handler_b", code: 200,
			params: Params{{"abc", "x/y/z"}}, pattern: "/foo/bar/*abc",
		}, {
			method: "GET", path: "/foo/x/y/z",
			handler: "handler_c", code: 200,
			params: Params{{"abc", "x/y/z"}}, pattern: "/foo/*abc",
		}, {
			method: "GET", path: "/x/y/z",
			handler: "handler_d", code: 200,
			params: Params{{"abc", "x/y/z"}}, pattern: "/*abc",
		}, {
			method: "GET", path: "/goo/car/x/y/z",
			handler: "handler_e", code: 200,
			params: Params{{"", "x/y/z"}}, pattern: "/goo/car/*",
		}, {
			method: "GET", path: "/goo/x/y/z",
			handler: "", code: 404,
			params: Params{}, pattern: "",
		}, {
			method: "GET", path: "/goo/xyz",
			handler: "handler_f", code: 200,
			params: Params{{"b", "xyz"}}, pattern: "/goo/{b}",
		},
	}.Run(t, router)
}

func TestRouterServeHTTP_Host(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar", "handler_a"},
		{"GET", "example.com/foo/bar", "handler_b"},
		{"GET", "{sub}.sample.{tld}/foo/bar", "handler_c"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar",
			handler: "handler_a", code: 200,
			params: Params{}, pattern: "/foo/bar",
		}, {
			method: "GET", path: "http://example.com/foo/bar",
			handler: "handler_b", code: 200,
			params: Params{}, pattern: "example.com/foo/bar",
		}, {
			method: "GET", path: "http://www.sample.co.uk/foo/bar",
			handler: "handler_c", code: 200,
			params: Params{{"sub", "www"}, {"tld", "co.uk"}}, pattern: "{sub}.sample.{tld}/foo/bar",
		}, {
			method: "GET", path: "http://www.example.com/foo/bar",
			handler: "handler_a", code: 200,
			params: Params{}, pattern: "/foo/bar",
		},
	}.Run(t, router)
}

func TestRouterServeHTTP_NotFound(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar", "handler"},
	}.Router()

	if router.handle404 == nil {
		t.Error("NewRouter() has not initialized the default not-found handler")
	}
	router.SetNotFound(nil)
	if router.handle404 == nil {
		t.Error("Router.SetNotFound(nil) should be a nop")
	}

	router.SetNotFound(HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Handled-By", r.URL.Path+" not found")
		w.WriteHeader(http.StatusNotFound)
		recordContext(c, w)
	}))

	routerTests{
		{
			method: "GET", path: "/", params: nil,
			handler: "/ not found", code: 404, pattern: "",
		}, {
			method: "GET", path: "/foo/baz", params: nil,
			handler: "/foo/baz not found", code: 404, pattern: "",
		}, {
			method: "GET", path: "/foo", params: nil,
			handler: "/foo not found", code: 404, pattern: "",
		},
	}.Run(t, router)

}

func TestRouterHandle_GET(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/bar", "handler_get"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar", handler: "handler_get",
			code: 200, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PUT", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "POST", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PATCH", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "DELETE", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		},
	}.Run(t, router)
}

func TestRouterHandle_POST(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"POST", "/foo/bar", "handler_post"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PUT", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "POST", path: "/foo/bar", handler: "handler_post",
			code: 200, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PATCH", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "DELETE", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		},
	}.Run(t, router)
}

func TestRouterHandle_PUT(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"PUT", "/foo/bar", "handler_put"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PUT", path: "/foo/bar", handler: "handler_put",
			code: 200, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "POST", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PATCH", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "DELETE", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		},
	}.Run(t, router)
}

func TestRouterHandle_PATCH(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"PATCH", "/foo/bar", "handler_patch"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PUT", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "POST", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PATCH", path: "/foo/bar", handler: "handler_patch",
			code: 200, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "DELETE", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		},
	}.Run(t, router)
}

func TestRouterHandle_DELETE(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"DELETE", "/foo/bar", "handler_delete"},
	}.Router()

	routerTests{
		{
			method: "GET", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PUT", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "POST", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "PATCH", path: "/foo/bar", handler: "",
			code: 405, params: Params{}, pattern: "/foo/bar",
		}, {
			method: "DELETE", path: "/foo/bar", handler: "handler_delete",
			code: 200, params: Params{}, pattern: "/foo/bar",
		},
	}.Run(t, router)
}

func TestRouterServeHTTP(t *testing.T) {
	//t.Skip()
	w := newRecorder()
	r := mustNewRequest("GET", "/", nil)
	r.RequestURI = "*"

	router := NewRouter()
	router.ServeHTTP(w, r)
	equals(t, 0, w.HeaderMap.Get("Connection"), "close")
	equals(t, 0, w.Code, http.StatusBadRequest)
}

func TestRouterHandleFunc(t *testing.T) {
	//t.Skip()
	router := routerSetup{
		{"GET", "/foo/*", "handler_foo"},
	}.Router()

	w := newRecorder()
	r := mustNewRequest("GET", "/foo/bar-baz-qux", nil)

	router.ServeHTTP(w, r)
	equals(t, 0, w.Params(), Params{{"", "bar-baz-qux"}})
	equals(t, 0, w.HeaderMap.Get("Handled-By"), "handler_foo")
}
