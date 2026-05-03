package middleware

import (
	kyerrors "kyrux/core/errors"
	"kyrux/core/router"
	"kyrux/core/security/auth"
	"net/http"
	"runtime/debug"
	"strings"
)

func AllowedHosts(hosts []string, debug bool) router.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(hosts))
	wildcard := false
	for _, h := range hosts {
		if h == "*" {
			wildcard = true
			break
		}
		allowed[h] = struct{}{}
	}
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) {
			if debug || wildcard {
				next(ctx)
				return
			}
			host := ctx.Request.Host
			if i := strings.LastIndex(host, ":"); i != -1 {
				host = host[:i]
			}
			if _, ok := allowed[host]; !ok {
				http.Error(ctx.Writer, "400 Bad Request — host não permitido", http.StatusBadRequest)
				return
			}
			next(ctx)
		}
	}
}

func RequireAuth(a *auth.Authenticator) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) {
			header := ctx.Request.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			claims, err := a.ValidateToken(token)
			if err != nil {
				ctx.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
				return
			}
			ctx.Set("claims", claims)
			next(ctx)
		}
	}
}

func CORS(allowedOrigins []string) router.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) {
			origin := ctx.Request.Header.Get("Origin")
			if _, ok := allowed[origin]; ok {
				ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if ctx.Request.Method == http.MethodOptions {
				ctx.Writer.WriteHeader(http.StatusNoContent)
				return
			}
			next(ctx)
		}
	}
}

func Recovery() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) {
			defer func() {
				if rec := recover(); rec != nil {
					kyerrors.RenderPanic(ctx.Writer, ctx.Request, rec, debug.Stack())
				}
			}()
			next(ctx)
		}
	}
}
