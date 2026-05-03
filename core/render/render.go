package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	kyerrors "kyrux/core/errors"
	"kyrux/core/router"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var templateFuncs = template.FuncMap{
	"url": router.Resolve,
}

const liveScript = `<script>
(function(){
  var p=location.protocol==="https:"?"wss":"ws";
  var ws=new WebSocket(p+"://"+location.host+"/kyrux/websocket/ws/");
  ws.onmessage=function(e){
    try{
      var m=JSON.parse(e.data);
      if(m.type!=="kyrux:dom")return;
      var el=document.querySelector('[kyrux-target="'+m.target+'"]');
      if(!el)return;
      var a=m.action||"replace";
      if(a==="append")el.insertAdjacentHTML("beforeend",m.html);
      else if(a==="prepend")el.insertAdjacentHTML("afterbegin",m.html);
      else if(a==="remove")el.remove();
      else el.innerHTML=m.html;
    }catch(_){}
  };
})();
</script>`

const reloadScript = `<script>
(function(){
  var es=new EventSource('/__kyrux_reload__');
  es.onmessage=function(e){if(e.data==='reload')location.reload();};
  es.onerror=function(){setTimeout(function(){location.reload();},1000);};
})();
</script>`

var (
	AppsDir           = "apps"
	defaultProcessors []ContextProcessor
	renderersMu       sync.RWMutex
	renderers         = map[string]*Renderer{}
	debugMode         atomic.Bool
)

func SetDebug(v bool)  { debugMode.Store(v) }
func isDebug() bool    { return debugMode.Load() }

func AddDefaultProcessor(p ContextProcessor) {
	defaultProcessors = append(defaultProcessors, p)
}

type ContextProcessor func(ctx *router.Context) map[string]any

type Renderer struct {
	engine     *Engine
	processors []ContextProcessor
}

// For retorna um Renderer cacheado para apps/<appName>/templates
// com os processadores padrão do framework já aplicados.
func For(appName string) *Renderer {
	renderersMu.RLock()
	if r, ok := renderers[appName]; ok {
		renderersMu.RUnlock()
		return r
	}
	renderersMu.RUnlock()

	renderersMu.Lock()
	defer renderersMu.Unlock()

	if r, ok := renderers[appName]; ok {
		return r
	}

	dir := filepath.Join(AppsDir, appName, "templates")
	r := &Renderer{
		engine:     MustNew(dir),
		processors: append([]ContextProcessor{}, defaultProcessors...),
	}
	renderers[appName] = r
	return r
}

func (r *Renderer) With(processors ...ContextProcessor) *Renderer {
	r.processors = append(r.processors, processors...)
	return r
}

func (r *Renderer) Render(ctx *router.Context, template string, data map[string]any) {
	setCurrentCtx(ctx)
	defer clearCurrentCtx()

	merged := mergedPool.Get().(map[string]any)
	for k := range merged {
		delete(merged, k)
	}
	for _, p := range r.processors {
		for k, v := range p(ctx) {
			merged[k] = v
		}
	}
	for k, v := range data {
		merged[k] = v
	}
	err := r.engine.Render(ctx.Writer, template, merged)
	mergedPool.Put(merged)
	if err != nil {
		log.Printf("render: %v", err)
		kyerrors.RenderPanic(ctx.Writer, ctx.Request, err, nil)
	}
}

type Engine struct {
	sources  map[string]srcInfo
	compiled map[string]*compiledEntry
	dir      string
	mu       sync.RWMutex
}

func New(dir string) (*Engine, error) {
	e := &Engine{dir: dir}
	return e, e.loadSources()
}

func (e *Engine) reload() error {
	return e.loadSources()
}

var (
	bufPool    = sync.Pool{New: func() any { return new(bytes.Buffer) }}
	mergedPool = sync.Pool{New: func() any { return make(map[string]any, 8) }}
)

func (e *Engine) Render(w http.ResponseWriter, name string, data any) error {
	if isDebug() {
		e.mu.Lock()
		reloadErr := e.reload()
		e.mu.Unlock()
		if reloadErr != nil {
			return reloadErr
		}
	}

	var ce *compiledEntry
	if isDebug() {
		e.mu.RLock()
		ce = e.compiled[name]
		e.mu.RUnlock()
	} else {
		ce = e.compiled[name]
	}

	if ce == nil {
		return fmt.Errorf("template '%s' não encontrado", name)
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	if err := ce.set.ExecuteTemplate(buf, ce.execName, data); err != nil {
		bufPool.Put(buf)
		return err
	}

	inject := liveScript
	if isDebug() {
		inject += reloadScript
	}
	s := strings.Replace(buf.String(), "</body>", inject+"</body>", 1)
	bufPool.Put(buf)
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Length", strconv.Itoa(len(s)))
	_, err := w.Write([]byte(s))
	return err
}

func (e *Engine) RenderToString(name string, data any) (string, error) {
	ce := e.compiled[name]
	if ce == nil {
		return "", fmt.Errorf("template '%s' não encontrado", name)
	}
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	err := ce.set.ExecuteTemplate(buf, ce.execName, data)
	s := buf.String()
	bufPool.Put(buf)
	return s, err
}

func (e *Engine) RenderFS(w http.ResponseWriter, fsys fs.FS, name string, data any) error {
	tmpl, err := template.ParseFS(fsys, "*.html")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, name, data)
}

func (e *Engine) RenderEmbed(w http.ResponseWriter, fsys fs.FS, name string, data any) error {
	var files []string
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && filepath.Ext(path) == ".html" {
			files = append(files, path)
		}
		return nil
	})
	tmpl, err := template.ParseFS(fsys, files...)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, name, data)
}

func StaticEmbedHandler(fsys fs.FS) http.Handler {
	return http.FileServer(http.FS(fsys))
}

func MustNew(dir string) *Engine {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return &Engine{dir: dir, sources: map[string]srcInfo{}, compiled: map[string]*compiledEntry{}}
	}
	e, err := New(dir)
	if err != nil {
		panic("render: " + err.Error())
	}
	return e
}
