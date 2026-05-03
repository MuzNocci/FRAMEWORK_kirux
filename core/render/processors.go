package render

import (
	"kyrux/core/environment"
	"kyrux/core/router"
	"runtime"
)

func AppContext(version string) ContextProcessor {
	cached := map[string]any{
		"AppName":   environment.GetOr("APP_NAME", "kyrux"),
		"Version":   version,
		"Env":       environment.GetOr("APP_ENV", "production"),
		"Addr":      environment.GetOr("SERVER_HOST", "0.0.0.0") + ":" + environment.GetOr("SERVER_PORT", "8080"),
		"GoVersion": runtime.Version(),
	}
	return func(_ *router.Context) map[string]any { return cached }
}
