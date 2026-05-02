package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"kyrux/core/environment"
	"kyrux/core/router"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const reloadScript = `<script>
(function(){
  var es = new EventSource('/__kyrux_reload__');
  es.onmessage = function(e){ if(e.data==='reload') location.reload(); };
  es.onerror   = function(){ setTimeout(function(){ location.reload(); }, 1000); };
})();
</script>`

var (
	AppsDir           = "apps"
	defaultProcessors []ContextProcessor
	renderersMu       sync.RWMutex
	renderers         = map[string]*Renderer{}
)

func isDebug() bool {
	return environment.GetOr("APP_DEBUG", "false") == "true"
}

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
	merged := make(map[string]any)
	for _, p := range r.processors {
		for k, v := range p(ctx) {
			merged[k] = v
		}
	}
	for k, v := range data {
		merged[k] = v
	}
	if err := r.engine.Render(ctx.Writer, template, merged); err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
	}
}

type Engine struct {
	templates map[string]*template.Template
	dir       string
	mu        sync.RWMutex
}

func New(dir string) (*Engine, error) {
	e := &Engine{dir: dir, templates: map[string]*template.Template{}}
	return e, e.reload()
}

func (e *Engine) reload() error {
	templates := map[string]*template.Template{}

	err := filepath.WalkDir(e.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".html" {
			return err
		}
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return err
		}
		name := filepath.Base(path)
		templates[name] = tmpl
		return nil
	})
	if err != nil {
		return err
	}

	e.templates = templates
	return nil
}

func (e *Engine) execute(name string, data any) (string, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		return "", fmt.Errorf("template '%s' não encontrado", name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *Engine) Render(w http.ResponseWriter, name string, data any) error {
	if isDebug() {
		e.mu.Lock()
		_ = e.reload()
		e.mu.Unlock()
	}

	e.mu.RLock()
	body, err := e.execute(name, data)
	e.mu.RUnlock()

	if err != nil {
		return err
	}

	if isDebug() {
		body = strings.Replace(body, "</body>", reloadScript+"</body>", 1)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = w.Write([]byte(body))
	return err
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
		return &Engine{dir: dir, templates: map[string]*template.Template{}}
	}
	e, err := New(dir)
	if err != nil {
		panic("render: " + err.Error())
	}
	return e
}
