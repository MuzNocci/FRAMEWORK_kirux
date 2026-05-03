package router

import (
	"encoding/json"
	kyerrors "kyrux/core/errors"
	"net/http"
	"strconv"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	data    map[string]any
}

func (c *Context) SetParam(key, value string) {
	if c.Params == nil {
		c.Params = make(map[string]string)
	}
	c.Params[key] = value
}

func (c *Context) Set(key string, value any) {
	if c.data == nil {
		c.data = make(map[string]any)
	}
	c.data[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	v, ok := c.data[key]
	return v, ok
}

func (c *Context) JSON(status int, v any) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	json.NewEncoder(c.Writer).Encode(v)
}

func (c *Context) HTML(status int, body string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(body))
}

func (c *Context) Redirect(url string, status int) {
	http.Redirect(c.Writer, c.Request, url, status)
}

func (c *Context) Error(code int) {
	kyerrors.Render(c.Writer, c.Request, code)
}

// Param retorna o valor do parâmetro de path pelo nome.
// Exemplo: rota "/usuarios/<id:int>/" → ctx.Param("id")
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// ParamInt retorna o parâmetro de path convertido para int.
// Retorna (0, false) se ausente ou não for um inteiro válido.
func (c *Context) ParamInt(key string) (int, bool) {
	v := c.Params[key]
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	return n, err == nil
}

// Query retorna o primeiro valor do parâmetro de query string.
// Exemplo: /busca?q=golang → ctx.Query("q")
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault retorna o parâmetro de query string ou fallback se ausente/vazio.
func (c *Context) QueryDefault(key, fallback string) string {
	if v := c.Request.URL.Query().Get(key); v != "" {
		return v
	}
	return fallback
}

// QueryInt retorna o parâmetro de query string como int, ou fallback se inválido.
func (c *Context) QueryInt(key string, fallback int) int {
	if n, err := strconv.Atoi(c.Request.URL.Query().Get(key)); err == nil {
		return n
	}
	return fallback
}

// QueryAll retorna todos os valores de um parâmetro de query string.
// Exemplo: /busca?tag=go&tag=web → ctx.QueryAll("tag") → ["go", "web"]
func (c *Context) QueryAll(key string) []string {
	return c.Request.URL.Query()[key]
}
