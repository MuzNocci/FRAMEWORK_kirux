package welcome

import (
	_ "embed"
	"html/template"
	"kyrux/core/environment"
	"kyrux/core/router"
	"runtime"
)

//go:embed welcome.html
var welcomeHTML string

type pageData struct {
	AppName   string
	Version   string
	Env       string
	Addr      string
	GoVersion string
}

var welcomeTpl = template.Must(template.New("welcome").Parse(welcomeHTML))

func RegisterIfNeeded(r *router.Router) {
	if r.HasRoute("GET /") {
		return
	}
	r.Handle("GET /", handler)
}

func handler(ctx *router.Context) {
	d := pageData{
		AppName:   environment.GetOr("APP_NAME", "kyrux"),
		Version:   environment.GetOr("APP_VERSION", "0.1.0"),
		Env:       environment.GetOr("APP_ENV", "production"),
		Addr:      environment.GetOr("SERVER_HOST", "0.0.0.0") + ":" + environment.GetOr("SERVER_PORT", "8080"),
		GoVersion: runtime.Version(),
	}
	ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = welcomeTpl.Execute(ctx.Writer, d)
}
