# KYRUX FRAMEWORK
Criado e desenvolvido por Müller Nocciolli.
Contato: muller.nocciolli@gmail.com
Site: www.nocciolli.com.br
Documentação: www.kyrux.com.br/docs/


## VISÃO GERAL:
Framework web em Go baseado em SSR, EventBus e Realtime invisível.


## PRINCÍPIOS:
- SSR-first
- Event-driven architecture
- Realtime automático e invisível
- Security centralizada
- Dev foca apenas na lógica


## CORE MODULES:

### Bootstrap:
Inicializa todo o framework (env, db, cache, router, realtime, csrf, allowed hosts)

### Environment:
Leitura de .env e variáveis do sistema com suporte a comentários inline

### Settings:
Configuração tipada global. Debug ligado automaticamente quando APP_ENV=development

### Router:
Mapeamento de URLs para views com resolução de nomes e rastreamento de rotas registradas

### Render:
Renderização SSR com herança de templates, sync.Pool de buffers e injeção automática de realtime

### Templates:
Herança e composição Django-like com extends, block, endblock e include

### Middleware:
Compressão gzip, CORS, recovery e cadeia de middlewares por rota

### Security:
CSRF automático, Allowed Hosts, autenticação JWT, sessão e criptografia

### EventBus:
Sistema interno de eventos desacoplados

### Realtime:
WebSocket invisível com atualização de DOM sem JS manual

### CLI:
Criação e remoção de apps via linha de comando


## COMANDOS:

### Iniciar o servidor:
```bash
go run main.go
```
O modo é detectado automaticamente pelo APP_ENV:
- `development` → inicia com Air (live reload). Requer Air instalado.
- `production` → inicia o servidor diretamente.

### Instalar o Air (uma vez):
```bash
go install github.com/air-verse/air@latest
```

### Criar um novo app:
```bash
go run main.go startapp <nome>
```

### Remover um app existente:
```bash
go run main.go removeapp <nome>
```


## VARIÁVEIS DE AMBIENTE (.env):

```env
# Environment
# development → debug, hotreload e pprof ativados automaticamente
# production  → modo otimizado, debug desligado
APP_ENV=development

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8000
SERVER_WORKERS=4  # omitir para usar todos os CPUs disponíveis

# Database
DB_ENABLED=true
DB_DRIVER=postgres
DB_DSN=postgres://user:password@localhost:5432/kyrux?sslmode=disable

# Cache
CACHE_ENABLED=false
CACHE_DRIVER=memory
CACHE_ADDR=localhost:6379

# Security
SECRET_KEY=your-strong-random-secret-key-here
SESSION_TTL=3600
ALLOWED_HOSTS=localhost,127.0.0.1  # omitir em development (ignorado automaticamente)
```


## TEMPLATES:

### Variáveis do framework (sem ponto):
```html
{{ AppName }}    {{ Version }}    {{ Env }}
{{ Addr }}       {{ GoVersion }}  {{ url "nome_da_rota" }}
```

### Variáveis da view (com ponto):
```html
{{ .titulo }}    {{ .usuario }}    {{ .produtos }}
```

A convenção é visual: sem ponto = framework, com ponto = seus dados.

### Herança de templates:

**base.html:**
```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>{% block "title" %}{{ AppName }}{% endblock "title" %}</title>
</head>
<body>
  {% block "content" %}{% endblock "content" %}
</body>
</html>
```

**página.html:**
```html
{% extends "base.html" %}

{% block "title" %}Minha Página{% endblock "title" %}

{% block "content" %}
  {% include "partials/header.html" %}
  <h1>{{ .titulo }}</h1>
{% endblock "content" %}
```

### CSRF em formulários:
```html
<form method="POST" action="{{ url "criar_usuario" }}">
  {{ csrf_token }}
  <input type="text" name="nome">
  <button type="submit">Salvar</button>
</form>
```
O CSRF é validado automaticamente em POST, PUT, PATCH e DELETE.
Para AJAX, envie o token no header `X-CSRF-Token`.


## REALTIME (DOM sem JS):

O Kyrux injeta automaticamente o WebSocket em toda página.
O desenvolvedor usa apenas atributos HTML e funções Go.

### Template:
```html
<div kyrux-target="lista-usuarios">
  {% include "partials/lista.html" %}
</div>
```

### View:
```go
func CriarUsuario(ctx *router.Context) {
    // ... salva no banco ...

    html, _ := render.Partial("users", "partials/lista.html", map[string]any{
        "usuarios": usuarios,
    })

    fw.Realtime.Replace("lista-usuarios", html)  // substitui o conteúdo
    fw.Realtime.Append("lista-usuarios", html)   // adiciona ao final
    fw.Realtime.Prepend("lista-usuarios", html)  // adiciona ao início
    fw.Realtime.Remove("lista-usuarios")         // remove o elemento
}
```

Zero JS escrito pelo desenvolvedor. O framework cuida de tudo.


## BANCO DE DADOS:

### Múltiplas conexões:
O Kyrux não importa nenhum driver — isso é responsabilidade do desenvolvedor.
Adicione o driver no `go.mod` e importe com blank identifier:

```go
import _ "github.com/lib/pq"          // PostgreSQL
import _ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL (pgx)
import _ "github.com/go-sql-driver/mysql"  // MySQL / MariaDB
import _ "modernc.org/sqlite"          // SQLite (sem CGO)
```

### Conexão padrão (via .env):
```env
DB_ENABLED=true
DB_DRIVER=postgres
DB_DSN=postgres://user:password@localhost:5432/kyrux?sslmode=disable
```

### Múltiplos bancos:
```go
fw.DB.Add("analytics", "postgres", os.Getenv("ANALYTICS_DSN"))
fw.DB.Add("legacy",    "mysql",    os.Getenv("LEGACY_DSN"))
```

### Uso:
```go
db := fw.DB.Use()             // conexão "default"
db := fw.DB.Use("analytics")  // conexão nomeada

rows, err := db.Query("SELECT id, nome FROM usuarios")
```

### Drivers suportados:
| Driver       | Módulo Go                              | Observação          |
|-------------|----------------------------------------|---------------------|
| postgres     | github.com/lib/pq                      |                     |
| pgx          | github.com/jackc/pgx/v5/stdlib         | melhor performance  |
| mysql        | github.com/go-sql-driver/mysql         | também MariaDB      |
| sqlite       | modernc.org/sqlite                     | sem CGO             |
| sqlite3      | github.com/mattn/go-sqlite3            | requer CGO          |
| sqlserver    | github.com/microsoft/go-mssqldb        |                     |
| oracle       | github.com/sijms/go-ora/v2             | puro Go, sem CGO    |

Clientes nativos (MongoDB, Redis, Cassandra, DynamoDB) não usam `database/sql` — consulte `core/database/drivers.go` para referência completa.


## URLS:

### Definição:
```go
var URLPatterns = []router.URLPattern{
    router.Path("GET",  "/",        views.HomeView,   "home"),
    router.Path("POST", "/usuarios", views.CriarUsuario, "criar_usuario"),
}
```

### Uso no template:
```html
<a href="{{ url "home" }}">Início</a>
<form action="{{ url "criar_usuario" }}">...</form>
```


## FLUXO DO SISTEMA:
Request → Allowed Hosts → CSRF → Middleware → Router → View → Service → DB
→ EventBus → Realtime → DOM atualizado automaticamente


## FLUXO DE DESENVOLVIMENTO:
Arquivo .go alterado    → Air detecta → Recompila → Reinicia servidor
Arquivo .html/.css/.js  → hotreload detecta → SSE → Browser recarrega


## PROFILING (apenas em development):
Disponível automaticamente em `http://localhost:6060/debug/pprof/`


## CONCEITO CHAVE:
Kyrux representa execução no "momento certo" (Kairos),
onde eventos dirigem a atualização do sistema em tempo real.


## LICENÇA:
MIT License com cláusula de atribuição obrigatória.
Qualquer projeto que utilize o Kyrux deve incluir créditos visíveis ao framework.
Veja o arquivo LICENSE para detalhes.
