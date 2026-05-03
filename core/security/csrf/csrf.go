package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"html/template"
	"kyrux/core/render"
	"kyrux/core/router"
	"net/http"
)

const (
	cookieName = "kyrux_csrf"
	fieldName  = "kyrux_csrf_token"
	headerName = "X-CSRF-Token"
	tokenLen   = 32
	contextKey = "csrf_token"
)

var unsafeMethods = map[string]bool{
	"POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}

var secureCookie bool

// SetSecure ativa a flag Secure no cookie CSRF. Deve ser chamado no bootstrap.
func SetSecure(v bool) { secureCookie = v }

// RegisterFuncs registra {{ csrf_token }} no FuncMap global de templates.
// Deve ser chamado no bootstrap antes do primeiro render.
func RegisterFuncs() {
	render.AddFunc("csrf_token", func() template.HTML {
		ctx := render.GetCurrentCtx()
		if ctx == nil {
			return ""
		}
		v, _ := ctx.Get(contextKey)
		token, _ := v.(string)
		return template.HTML(`<input type="hidden" name="` + fieldName + `" value="` + token + `">`)
	})
}

func Middleware(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) {
		token := getOrCreate(ctx)
		ctx.Set(contextKey, token)

		if unsafeMethods[ctx.Request.Method] {
			submitted := ctx.Request.FormValue(fieldName)
			if submitted == "" {
				submitted = ctx.Request.Header.Get(headerName)
			}
			if !valid(token, submitted) {
				http.Error(ctx.Writer, "403 Forbidden — CSRF token inválido ou ausente", http.StatusForbidden)
				return
			}
		}

		next(ctx)
	}
}

func getOrCreate(ctx *router.Context) string {
	if c, err := ctx.Request.Cookie(cookieName); err == nil && c.Value != "" {
		return c.Value
	}
	token := generate()
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: false,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	return token
}

func generate() string {
	b := make([]byte, tokenLen)
	if _, err := rand.Read(b); err != nil {
		panic("csrf: falha ao gerar token")
	}
	return hex.EncodeToString(b)
}

func valid(expected, submitted string) bool {
	if len(submitted) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(submitted)) == 1
}
