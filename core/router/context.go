package router

import (
	"encoding/json"
	"net/http"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	data    map[string]any
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
