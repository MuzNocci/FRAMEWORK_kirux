package middleware

import (
	"compress/gzip"
	"kyrux/core/router"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(nil) }}

type gzipWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (gw *gzipWriter) Write(b []byte) (int, error) {
	gw.Header().Del("Content-Length")
	return gw.gz.Write(b)
}

// Compress é um middleware que comprime respostas com gzip quando o cliente suporta.
func Compress(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) {
		if !strings.Contains(ctx.Request.Header.Get("Accept-Encoding"), "gzip") {
			next(ctx)
			return
		}
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(ctx.Writer)
		defer func() {
			gz.Close()
			gzipPool.Put(gz)
		}()
		ctx.Writer.Header().Set("Content-Encoding", "gzip")
		ctx.Writer = &gzipWriter{ctx.Writer, gz}
		next(ctx)
	}
}
