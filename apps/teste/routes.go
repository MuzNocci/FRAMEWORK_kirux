package teste

import (
	"kyrux/apps/teste/views"
	"kyrux/core/bootstrap"
	"kyrux/core/router"
)

func init() {
	bootstrap.RegisterApp("teste", Register)
}

var URLPatterns = []router.URLPattern{
	router.Path("GET", "/", views.ExemploView, "exemplo_home"),
}

func Register(r *router.Router) {
	router.Include(r, URLPatterns)
}
