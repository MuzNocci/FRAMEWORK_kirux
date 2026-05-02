package router

type URLPattern struct {
	Method string
	Path   string
	View   HandlerFunc
	Name   string
}

func Path(method, path string, view HandlerFunc, name string) URLPattern {
	return URLPattern{Method: method, Path: path, View: view, Name: name}
}

func Include(r *Router, patterns []URLPattern) {
	for _, p := range patterns {
		r.Handle(p.Method+" "+p.Path, p.View)
	}
}
