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
- Realtime automático
- Security centralizada
- Dev foca apenas na lógica


## CORE MODULES:

### Bootstrap:
Inicializa todo o framework (env, db, cache, router, realtime)

### Environment:
Leitura de .env e variáveis do sistema com suporte a comentários inline

### Settings:
Configuração tipada global com controle de workers e GOGC via .env

### Router:
Mapeamento de URLs para views com rastreamento de rotas registradas

### Render:
Renderização SSR com injeção automática de realtime, sync.Pool de buffers e Content-Length

### Middleware:
Compressão gzip e cadeia de middlewares por rota

### Security:
Autenticação, sessão, criptografia e middleware

### EventBus:
Sistema interno de eventos desacoplados

### Realtime:
WebSocket invisível + bridge com EventBus

### CLI:
Gerenciamento de apps e servidor de desenvolvimento via comandos


## COMANDOS:

### Iniciar o servidor:
```bash
go run main.go
```

### Iniciar em modo desenvolvimento com live reload (requer Air):
```bash
# instalar o Air (uma vez)
go install github.com/air-verse/air@latest

go run main.go dev
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
# Enviroment
# development → debug automático ligado
# production → debug automático desligado
APP_ENV=development

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8000
SERVER_WORKERS=4  # Quando não definido, serão utilizados todos os workers disponíveis no CPU.

# Database
DB_DRIVER=postgres
DB_DSN=postgres://user:password@localhost:5432/kyrux?sslmode=disable

# Cache
CACHE_DRIVER=memory
CACHE_ADDR=localhost:6379

# Security
SECRET_KEY=your-strong-random-secret-key-here
SESSION_TTL=3600
```


## FLUXO DO SISTEMA:
Request -> Middleware -> Router -> View -> Service -> DB
-> EventBus -> Realtime -> Client update automático


## FLUXO DE DESENVOLVIMENTO (com Air):
Arquivo .go alterado -> Air detecta -> Recompila -> Reinicia servidor
Arquivo .html/.css/.js alterado -> core/hotreload detecta -> SSE -> Browser recarrega


## CONCEITO CHAVE:
Kyrux representa execução no "momento certo" (Kairos),
onde eventos dirigem a atualização do sistema em tempo real.


## LICENÇA:
MIT License com cláusula de atribuição obrigatória.
Qualquer projeto que utilize o Kyrux deve incluir créditos visíveis ao framework.
Veja o arquivo LICENSE para detalhes.
