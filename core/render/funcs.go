package render

import (
	"bytes"
	"kyrux/core/router"
	"runtime"
	"strconv"
	"sync"
)

var goroutineCtxStore sync.Map

func gid() uint64 {
	var buf [32]byte
	n := runtime.Stack(buf[:], false)
	s := buf[10:n] // skip "goroutine "
	i := bytes.IndexByte(s, ' ')
	id, _ := strconv.ParseUint(string(s[:i]), 10, 64)
	return id
}

func setCurrentCtx(ctx *router.Context) { goroutineCtxStore.Store(gid(), ctx) }
func clearCurrentCtx()                  { goroutineCtxStore.Delete(gid()) }

// GetCurrentCtx retorna o *router.Context da goroutine em execução.
// Usado por funções do FuncMap que precisam de dados por request.
func GetCurrentCtx() *router.Context {
	if v, ok := goroutineCtxStore.Load(gid()); ok {
		return v.(*router.Context)
	}
	return nil
}

// AddFunc registra uma função no FuncMap global de templates.
// Deve ser chamado antes do primeiro render.
func AddFunc(name string, fn any) {
	templateFuncs[name] = fn
}
