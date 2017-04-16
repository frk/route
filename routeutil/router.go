package routeutil

import (
	"github.com/frk/route"
)

type Router struct {
	*route.Router
}

func NewRouter() *Router {
	return &Router{route.NewRouter()}
}
