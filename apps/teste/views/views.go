package views

import (
	"kyrux/core/render"
	"kyrux/core/router"
)

func ExemploView(ctx *router.Context) {
	context := map[string]any{
		"exemplo": "exemplo de renderização de template",
	}
	render.For("teste").Render(ctx, "exemplo.html", context)
}
