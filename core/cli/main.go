package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const installedFile = "core/apps/installed.go"
const settingsFile = "core/settings.go"

func Run(args []string) {
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "startapp", "removeapp":
		if len(args) < 2 {
			printUsage()
			os.Exit(1)
		}
		appName := strings.ToLower(args[1])
		switch command {
		case "startapp":
			if err := startApp(appName); err != nil {
				fmt.Fprintf(os.Stderr, "erro: %v\n", err)
				os.Exit(1)
			}
		case "removeapp":
			if err := removeApp(appName); err != nil {
				fmt.Fprintf(os.Stderr, "erro: %v\n", err)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "comando desconhecido: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("uso: go run main.go <comando> [app]")
	fmt.Println()
	fmt.Println("comandos:")
	fmt.Println("  startapp  <nome>   cria um novo app")
	fmt.Println("  removeapp <nome>   remove um app existente")
}

// ── startapp ─────────────────────────────────────────────────────────────────

func startApp(name string) error {
	base := filepath.Join("apps", name)

	if _, err := os.Stat(base); err == nil {
		return fmt.Errorf("app '%s' já existe em %s", name, base)
	}

	dirs := []string{
		filepath.Join(base, "statics", "css"),
		filepath.Join(base, "statics", "js"),
		filepath.Join(base, "templates"),
		filepath.Join(base, "views"),
		filepath.Join(base, "models"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("criar diretório %s: %w", dir, err)
		}
	}

	files := []struct {
		path    string
		content string
	}{
		{filepath.Join(base, "routes.go"), routesTpl},
		{filepath.Join(base, "views", "views.go"), viewsTpl},
		{filepath.Join(base, "models", "models.go"), modelsTpl},
		{filepath.Join(base, "templates", "exemplo.html"), templateTpl},
	}

	data := struct{ Name string }{Name: name}

	for _, f := range files {
		if err := writeTemplate(f.path, f.content, data); err != nil {
			return err
		}
	}

	if err := addInstalledImport(name); err != nil {
		return fmt.Errorf("atualizar %s: %w", installedFile, err)
	}

	if err := addInstalledApp(name); err != nil {
		return fmt.Errorf("atualizar %s: %w", settingsFile, err)
	}

	fmt.Printf("app '%s' criado em %s\n", name, base)
	return nil
}

// ── removeapp ────────────────────────────────────────────────────────────────

func removeApp(name string) error {
	base := filepath.Join("apps", name)

	if _, err := os.Stat(base); os.IsNotExist(err) {
		return fmt.Errorf("app '%s' não encontrado em %s", name, base)
	}

	fmt.Printf("remover o app '%s' em %s? essa ação é irreversível. [s/N] ", name, base)

	var input string
	fmt.Scanln(&input)

	if strings.ToLower(strings.TrimSpace(input)) != "s" {
		fmt.Println("operação cancelada.")
		return nil
	}

	if err := os.RemoveAll(base); err != nil {
		return fmt.Errorf("remover %s: %w", base, err)
	}

	if err := removeInstalledImport(name); err != nil {
		return fmt.Errorf("atualizar %s: %w", installedFile, err)
	}

	if err := removeInstalledApp(name); err != nil {
		return fmt.Errorf("atualizar %s: %w", settingsFile, err)
	}

	fmt.Printf("app '%s' removido.\n", name)
	return nil
}

// ── installed.go ─────────────────────────────────────────────────────────────

func addInstalledImport(name string) error {
	line := fmt.Sprintf("\t_ \"kyrux/apps/%s\"", name)
	content, err := os.ReadFile(installedFile)
	if err != nil {
		return err
	}
	if !strings.Contains(string(content), "import (") {
		block := "\nimport (\n" + line + "\n)\n"
		return os.WriteFile(installedFile, append(bytes.TrimRight(content, "\n"), []byte(block)...), 0644)
	}
	return addLineBeforeClosing(installedFile, ")", line)
}

func removeInstalledImport(name string) error {
	line := fmt.Sprintf("\t_ \"kyrux/apps/%s\"", name)
	if err := removeLine(installedFile, line); err != nil {
		return err
	}
	return collapseImportIfEmpty(installedFile)
}

// collapseImportIfEmpty remove o bloco import() quando ele ficar vazio.
func collapseImportIfEmpty(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "import (" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			t := strings.TrimSpace(lines[j])
			if t == "" {
				continue
			}
			if t == ")" {
				start := i
				if start > 0 && strings.TrimSpace(lines[start-1]) == "" {
					start--
				}
				result := append(lines[:start:start], lines[j+1:]...)
				return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
			}
			break
		}
	}
	return nil
}

// ── settings.go ──────────────────────────────────────────────────────────────

func addInstalledApp(name string) error {
	content, err := os.ReadFile(settingsFile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	entry := fmt.Sprintf("\t\t\t\"%s\",", name)

	for i, l := range lines {
		if !strings.Contains(l, "InstalledApps:") {
			continue
		}
		// Formato compacto: InstalledApps: []string{},
		if strings.Contains(l, "[]string{}") {
			indent := l[:strings.Index(l, "InstalledApps")]
			expanded := []string{indent + "InstalledApps: []string{", entry, indent + "},"}
			result := append(lines[:i:i], append(expanded, lines[i+1:]...)...)
			return os.WriteFile(settingsFile, []byte(strings.Join(result, "\n")), 0644)
		}
		// Formato expandido: encontra o }, de fechamento
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "}," {
				lines = append(lines[:j:j], append([]string{entry}, lines[j:]...)...)
				return os.WriteFile(settingsFile, []byte(strings.Join(lines, "\n")), 0644)
			}
		}
	}
	return fmt.Errorf("InstalledApps não encontrado em %s", settingsFile)
}

func removeInstalledApp(name string) error {
	entry := fmt.Sprintf("\t\t\t\"%s\",", name)
	if err := removeLine(settingsFile, entry); err != nil {
		return err
	}
	return collapseInstalledAppsIfEmpty(settingsFile)
}

// collapseInstalledAppsIfEmpty converte InstalledApps: []string{\n},  de volta para []string{},
func collapseInstalledAppsIfEmpty(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	for i, l := range lines {
		if !strings.Contains(l, "InstalledApps:") || !strings.Contains(l, "[]string{") {
			continue
		}
		if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "}," {
			indent := l[:strings.Index(l, "InstalledApps")]
			result := append(lines[:i:i], append([]string{indent + "InstalledApps: []string{},"}, lines[i+2:]...)...)
			return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
		}
	}
	return nil
}

// ── helpers de arquivo ────────────────────────────────────────────────────────

// addLineBeforeClosing insere newLine antes da última ocorrência de closing.
func addLineBeforeClosing(path, closing, newLine string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == strings.TrimSpace(closing) {
			lines = append(lines[:i], append([]string{newLine}, lines[i:]...)...)
			break
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// addLineAfterAnchor encontra a linha que contém anchor, depois insere
// newLine antes da primeira ocorrência de closing após essa linha.
func addLineAfterAnchor(path, anchor, closing, newLine string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	anchorIdx := -1
	for i, l := range lines {
		if strings.Contains(l, anchor) {
			anchorIdx = i
			break
		}
	}
	if anchorIdx == -1 {
		return fmt.Errorf("âncora '%s' não encontrada em %s", anchor, path)
	}

	for i := anchorIdx + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == strings.TrimSpace(closing) {
			lines = append(lines[:i], append([]string{newLine}, lines[i:]...)...)
			break
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func removeLine(path, target string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	result := lines[:0]
	for _, l := range lines {
		if strings.TrimSpace(l) != strings.TrimSpace(target) {
			result = append(result, l)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// ── helpers de template ───────────────────────────────────────────────────────

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func writeTemplate(path, content string, data any) error {
	tpl, err := template.New("").Funcs(template.FuncMap{"title": title}).Parse(content)
	if err != nil {
		return fmt.Errorf("parse template %s: %w", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("criar arquivo %s: %w", path, err)
	}
	defer f.Close()

	return tpl.Execute(f, data)
}

// ── templates dos arquivos gerados ───────────────────────────────────────────

var routesTpl = `package {{.Name}}

import (
	"kyrux/apps/{{.Name}}/views"
	"kyrux/core/bootstrap"
	"kyrux/core/router"
)

func init() {
	bootstrap.RegisterApp("{{.Name}}", Register)
}

var URLPatterns = []router.URLPattern{
	router.Path("GET", "/", views.ExemploView, "exemplo_home"),
}

func Register(r *router.Router) {
	router.Include(r, URLPatterns)
}
`

var viewsTpl = `package views

import (
	"kyrux/core/render"
	"kyrux/core/router"
)

func ExemploView(ctx *router.Context) {
	context := map[string]any{
		"example": "example",
	}
	render.For("{{.Name}}").Render(ctx, "exemplo.html", context)
}
`

var modelsTpl = `package models
`

var templateTpl = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{"{{"}} AppName {{"}}"}}</title>
<style>*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}:root{--go-blue:#00ACD7;--go-blue-light:#5DC9E2;--go-blue-dark:#00758D;--bg:#0D1117;--surface:#161B22;--border:#1E2A38;--text:#E6EDF3;--muted:#8B949E}html,body{height:100%}body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:var(--bg);color:var(--text);display:flex;flex-direction:column;overflow:auto}header{border-bottom:1px solid var(--border);padding:.9rem 2.5rem;display:flex;align-items:center;gap:.75rem;flex-shrink:0}.logo-mark{width:28px;height:28px;background:var(--go-blue);border-radius:6px;display:flex;align-items:center;justify-content:center;font-weight:800;font-size:.9rem;color:#fff;letter-spacing:-.5px}header span{font-size:1rem;font-weight:600;color:var(--text)}header .version{margin-left:auto;font-size:.7rem;color:var(--muted);background:var(--border);padding:.2rem .55rem;border-radius:999px}main{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:2.5rem 2rem;text-align:center;gap:1.1rem}.badge{display:inline-flex;align-items:center;gap:.4rem;font-size:.75rem;color:var(--go-blue-light);background:rgba(0,172,215,.1);border:1px solid rgba(0,172,215,.2);padding:.3rem .8rem;border-radius:999px;margin-bottom:.25rem}.dot{width:6px;height:6px;border-radius:50%;background:var(--go-blue-light);animation:pulse 2s ease-in-out infinite}@keyframes pulse{0%,100%{opacity:1}50%{opacity:.35}}h1{font-size:clamp(1.4rem,3.5vw,2.2rem);font-weight:800;letter-spacing:-.03em;color:var(--text)}.subtitle{font-size:clamp(.82rem,1.4vw,.95rem);color:var(--muted);max-width:420px;line-height:1.7}.cards{display:flex;gap:1rem;margin-top:.5rem;flex-wrap:wrap;justify-content:center}.card{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:1.1rem 1.4rem;text-align:left;width:180px;transition:border-color .2s}.card:hover{border-color:var(--go-blue-dark)}.card-icon{font-size:1.3rem;margin-bottom:.5rem}.card-title{font-size:.82rem;font-weight:700;color:var(--text);margin-bottom:.2rem}.card-desc{font-size:.75rem;color:var(--muted);line-height:1.5}.actions{display:flex;gap:.75rem;margin-top:.5rem;flex-wrap:wrap;justify-content:center}.btn{display:inline-flex;align-items:center;gap:.35rem;padding:.5rem 1.2rem;border-radius:8px;font-size:.82rem;text-decoration:none;font-weight:600;transition:border-color .2s,color .2s,background .2s}.btn-primary{background:var(--go-blue);color:#fff;border:1px solid var(--go-blue)}.btn-primary:hover{background:var(--go-blue-dark);border-color:var(--go-blue-dark)}.btn-ghost{background:transparent;border:1px solid var(--border);color:var(--muted)}.btn-ghost:hover{border-color:var(--go-blue-dark);color:var(--text)}footer{border-top:1px solid var(--border);padding:.75rem 2rem;text-align:center;font-size:.82rem;color:var(--muted);flex-shrink:0}footer a{color:var(--go-blue);text-decoration:none}footer a:hover{text-decoration:underline}@media(max-width:640px){main{padding:2rem 1.25rem;gap:1rem}.cards{flex-direction:column;align-items:center}.card{width:100%;max-width:280px}}@media(max-width:380px){.actions{flex-direction:column;align-items:center}}</style>
</head>
<body>
<header>
  <div class="logo-mark">K</div>
  <span>{{"{{"}} AppName {{"}}"}}</span>
  <span class="version">v{{"{{"}} Version {{"}}"}}</span>
</header>
<main>
  <span class="badge"><span class="dot"></span> servidor rodando</span>
  <h1>Bem-vindo ao {{"{{"}} AppName {{"}}"}}</h1>
  <p class="subtitle">Seu projeto está funcionando. Edite as views, rotas e templates para começar a construir.</p>

  <div class="cards">
    <div class="card">
      <div class="card-icon">&#128218;</div>
      <div class="card-title">Rotas</div>
      <div class="card-desc">Defina URLs em <code>routes.go</code> com <code>router.Path()</code>.</div>
    </div>
    <div class="card">
      <div class="card-icon">&#128065;</div>
      <div class="card-title">Views</div>
      <div class="card-desc">Lógica das páginas em <code>views/views.go</code> via <code>ctx</code>.</div>
    </div>
    <div class="card">
      <div class="card-icon">&#127912;</div>
      <div class="card-title">Templates</div>
      <div class="card-desc">HTML em <code>templates/</code> com o motor nativo do Go.</div>
    </div>
  </div>

  <div class="actions">
    <a href="/kyrux/debug/" class="btn btn-primary">&#9881; Debug</a>
    <a href="https://www.kyrux.com.br/docs/" target="_blank" class="btn btn-ghost">Documentação &#8599;</a>
  </div>
</main>
<footer>
  Kyrux Framework &mdash; construído com <a href="https://go.dev" target="_blank">Go</a>
  &nbsp;&middot;&nbsp;
  <a href="https://www.kyrux.com.br/docs/" target="_blank">Documentação</a>
  &nbsp;&middot;&nbsp;
  desenvolvido por <a href="https://www.nocciolli.com.br" target="_blank">M&uuml;ller Nocciolli</a>
</footer>
</body>
</html>
`
