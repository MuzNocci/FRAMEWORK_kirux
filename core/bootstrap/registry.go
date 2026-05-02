package bootstrap

import "kyrux/core/router"

type RegisterFunc func(r *router.Router)

var registry = map[string]RegisterFunc{}

func RegisterApp(name string, fn RegisterFunc) {
	registry[name] = fn
}
