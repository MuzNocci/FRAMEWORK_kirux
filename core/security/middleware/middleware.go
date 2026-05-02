package middleware

import (
	"kyrux/core/router"
	"kyrux/core/security/auth"
	"net/http"
	"strings"
)

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
				if r := recover(); r != nil {
					ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
				}
			}()
			next(ctx)
		}
	}
}
