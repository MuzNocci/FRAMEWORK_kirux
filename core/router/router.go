package router

import "net/http"

type HandlerFunc func(ctx *Context)

type Router struct {
	mux         *http.ServeMux
	middlewares []MiddlewareFunc
	routes      map[string]bool
}

type MiddlewareFunc func(HandlerFunc) HandlerFunc

func New() *Router {
	return &Router{mux: http.NewServeMux(), routes: map[string]bool{}}
}

func (r *Router) HasRoute(pattern string) bool {
	return r.routes[pattern]
}

func (r *Router) Use(m MiddlewareFunc) {
	r.middlewares = append(r.middlewares, m)
}

func (r *Router) Handle(pattern string, h HandlerFunc) {
	r.routes[pattern] = true
	chain := r.chain(h)
	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		chain(&Context{Writer: w, Request: req, Params: map[string]string{}})
	})
}

func (r *Router) chain(h HandlerFunc) HandlerFunc {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

func (r *Router) HandlePrefix(prefix string, h http.Handler) {
	r.mux.Handle(prefix, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
