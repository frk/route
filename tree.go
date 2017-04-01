package route

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type paramNode struct {
	start   byte
	end     byte
	name    string
	pattern string
	handler nodeHandler
	child   *node
}

type catchallNode struct {
	name    string
	pattern string
	handler nodeHandler
}

type node struct {
	edge      string
	pattern   string
	handler   nodeHandler
	maxParams uint8

	indices string

	children []*node
	param    *paramNode
	catchall *catchallNode
}

func (nd *node) insert(method, pattern string, h Handler) error {
	var (
		cn        = nd // current node
		pat       = pattern
		maxParams = countParams(pattern)
	)

Loop:
	for {
		if pat == "" {
			cn.pattern = pattern
			return cn.handler.set(method, h)
		}

		if maxParams > cn.maxParams {
			cn.maxParams = maxParams
		}

		// catch-all node
		if pat[0] == '*' {
			if cn.catchall == nil {
				cn.catchall = &catchallNode{}
			}
			if err := cn.catchall.handler.set(method, h); err != nil {
				return err
			}
			cn.catchall.name = pat[1:]
			cn.catchall.pattern = pattern
			break
		}

		// parameter node
		if pat[0] == '{' {
			i := strings.IndexByte(pat, '}')
			if i == -1 {
				return &routeError{typ: errUnclosedParam}
			}
			name := pat[1:i]

			var start, end byte
			if len(cn.edge) > 0 {
				start = cn.edge[len(cn.edge)-1]
			}
			if len(pat) > (i + 1) {
				end = pat[i+1]
			}

			if cn.param == nil {
				cn.param = &paramNode{name: name}
			}

			if cn.param.name != "" && cn.param.name != name {
				return &routeError{errParamConflict, name, cn.param.name}
			}
			if start != cn.param.start {
				if start != 0 && cn.param.start != 0 {
					return &routeError{errSeparatorConflict, start, cn.param.start}
				}
				if start == 0 {
					start = cn.param.start
				}
			}
			if end != cn.param.end {
				if end != 0 && cn.param.end != 0 {
					return &routeError{errSeparatorConflict, end, cn.param.end}
				}
				if end == 0 {
					end = cn.param.end
				}
			}

			cn.param.start = start
			cn.param.end = end
			cn.param.name = name

			pat = pat[i+1:]
			if pat == "" {
				cn.param.pattern = pattern
				return cn.param.handler.set(method, h)
			} else if cn.param.child == nil {
				cn.param.child = &node{}
			}

			maxParams--
			cn = cn.param.child
			continue Loop
		}

		// static node
		for i, n := range cn.children {
			if n.edge != "" && n.edge[0] == pat[0] {
				pl := cpl(n.edge, pat)
				if pl < len(n.edge) {
					// split the edge
					prefix, suffix := n.edge[:pl], n.edge[pl:]
					n = &node{
						edge:    prefix,
						indices: string([]byte{suffix[0]}),
						children: []*node{{
							edge:      suffix,
							pattern:   n.pattern,
							indices:   n.indices,
							maxParams: n.maxParams,
							handler:   n.handler,
							children:  n.children,
							param:     n.param,
							catchall:  n.catchall,
						}},
					}

					cn.children[i] = n
				}

				cn = n
				pat = pat[pl:]

				continue Loop
			}
		}

		var edge string
		if i := strings.IndexByte(pat, '{'); i != -1 {
			edge, pat = pat[:i], pat[i:]
		} else if i := strings.IndexByte(pat, '*'); i != -1 {
			edge, pat = pat[:i], pat[i:]
		} else {
			edge, pat = pat, ""
		}

		n := &node{
			edge:      edge,
			maxParams: maxParams,
		}
		cn.indices += string([]byte{edge[0]})
		cn.children = append(cn.children, n)
		cn = n
	}

	return nil
}

func (nd *node) lookup(path string, po Params) (h Handler, ps Params, pat string, redir tsr) {
	ps = po[0:0]

	var prev *node
	var pn, cn *node
	var pp, cp string

Loop:
	for {
		if path == "" || path == nd.edge {
			if nd.handler.isSet {
				pat = nd.pattern
				h = &nd.handler
				return
			}

			if nd.edge == "/" && (prev != nil && prev.handler.isSet) {
				return nil, nil, "", tsrWithoutSlash
			}
			return recommend(nd, path)
		}

		if nd.catchall != nil {
			pn = nil
			cn = nd
			cp = path
		}
		if nd.param != nil {
			cn = nil
			pn = nd
			pp = path
		}

		// static node
		c := path[0]
		for i := 0; i < len(nd.indices); i++ {
			if c == nd.indices[i] {
				n := nd.children[i]
				if plen, elen := len(path), len(n.edge); plen >= elen && n.edge == path[:elen] {
					path = path[elen:]
				} else {
					break
				}

				prev = nd
				nd = n
				continue Loop
			}
		}

		// parameter node
		if pn != nil {
			path = pp
			elen := len(pn.edge)
			if (elen == 0 && pn.param.start == 0) || (elen > 0 && pn.edge[elen-1] == pn.param.start) {
				var i int
				for plen := len(path); i < plen && (path[i] != pn.param.end && path[i] != '/'); i++ {
				}

				ps = append(ps, param{
					key: pn.param.name,
					val: path[:i],
				})

				path = path[i:]
				if path == "" {
					if pn.param.handler.isSet {
						pat = pn.param.pattern
						h = &pn.param.handler
						return
					}
					return recommend(pn.param.child, path)
				} else if pn.param.child == nil {
					if path == "/" && pn.param.handler.isSet {
						return nil, nil, "", tsrWithoutSlash
					}
					return nil, nil, "", tsrNone
				}

				prev = pn
				nd = pn.param.child
				pn = nil
				cn = nil
				continue
			}
		}

		// catch-all node
		if cn != nil {
			path = cp
			ps = append(ps, param{
				key: cn.catchall.name,
				val: path,
			})

			if !cn.catchall.handler.isSet {
				return nil, nil, "", tsrNone
			}
			pat = cn.catchall.pattern
			h = &cn.catchall.handler
		}

		break Loop
	}

	if h == nil {
		if path == "/" && nd.handler.isSet {
			return nil, nil, "", tsrWithoutSlash
		}
		return recommend(nd, path)
	}
	return
}

func recommend(nd *node, path string) (h Handler, ps Params, pat string, redir tsr) {
	if plen := len(path); plen == 0 || path[plen-1] != '/' {
		path += "/"
		if nd != nil {
			for _, n := range nd.children {
				if n.edge == path && n.handler.isSet {
					return nil, nil, "", tsrWithSlash
				}
			}
		}
	}
	return nil, nil, "", tsrNone
}

func countParams(pattern string) (n uint8) {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			return n + 1
		} else if pattern[i] == '{' {
			n++
		}
	}
	return n
}

func cpl(a, b string) int {
	var i int
	for j := min(len(a), len(b)); i < j; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type nodeHandler struct {
	isSet bool // isSet reports whether at least one Handler is set in the hm field.

	// The hm field is a map that associates Handlers with http methods.
	hm map[string]Handler

	// The methods field contains a string of lexicographically sorted comma
	// separated http methods that can be handled by the node.
	methods string
}

// ServeHTTP implements the route.Handler interface.
func (nh *nodeHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h := nh.hm[r.Method]
	if h == nil {
		h = nh.hm["*"]
	}
	if h == nil {
		w.Header().Set("Allow", nh.methods)
		if r.Method != "OPTIONS" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		h.ServeHTTP(c, w, r)
	}
}

func (nh *nodeHandler) set(method string, h Handler) error {
	if nh.hm == nil {
		nh.hm = map[string]Handler{}
	}

	ms := strings.Split(method, ",")
	for _, m := range ms {
		if m == "" {
			return fmt.Errorf("Missing method")
		}
		if _, ok := nh.hm[m]; ok {
			return &routeError{typ: errMethodConflict, a: m}
		}
		nh.hm[m] = h
	}
	nh.isSet = true

	// On each call to "set" re-iterate over all methods, sort them and
	// set the resulting value to the methods field.
	var methods []string
	for m, _ := range nh.hm {
		if m != "*" {
			methods = append(methods, m)
		}
	}
	sort.Strings(methods)
	nh.methods = strings.Join(methods, ",")

	return nil
}

type errorType int

const (
	errUnclosedParam errorType = iota
	errParamConflict
	errSeparatorConflict
	errMethodConflict
)

type routeError struct {
	typ  errorType
	a, b interface{} // values that caused the error
}

func (e *routeError) Error() string {
	switch e.typ {
	case errUnclosedParam:
		return "missing closing curly brace '}'"
	case errParamConflict:
		return fmt.Sprintf("The param name %q conflicts with the param "+
			"name %q in the same segment of a previously registered pattern.", e.a, e.b)
	case errSeparatorConflict:
		return fmt.Sprintf("The param separator '%c' conflicts with the "+
			"separator '%c' in the same location of a previously registered pattern.", e.a, e.b)
	case errMethodConflict:
		return fmt.Sprintf("A handler for the %q method is already registered.", e.a)
	default:
		return "unknown error"
	}
}
