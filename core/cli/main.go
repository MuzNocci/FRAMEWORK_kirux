package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
	case "dev":
		runDev()
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
	fmt.Println("  dev                inicia o servidor em modo desenvolvimento com Air")
	fmt.Println("  startapp  <nome>   cria um novo app")
	fmt.Println("  removeapp <nome>   remove um app existente")
}

func runDev() {
	if _, err := exec.LookPath("air"); err != nil {
		fmt.Fprintln(os.Stderr, "Air não encontrado. Instale com:")
		fmt.Fprintln(os.Stderr, "  go install github.com/air-verse/air@latest")
		os.Exit(1)
	}
	fmt.Println("Iniciando Kyrux em modo desenvolvimento com Air...")
	cmd := exec.Command("air", "-c", "core/air/.air.toml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
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
		{filepath.Join(base, "templates", "index.html"), templateTpl},
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
	router.Path("GET", "", views.{{title .Name}}View, "url_name"),
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

func {{title .Name}}View(ctx *router.Context) {
	context := map[string]any{
		"example": "example",
	}
	render.For("{{.Name}}").Render(ctx, "index.html", context)
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
</head>
<body>
  <h1>{{.Name}}</h1>
</body>
</html>
`
