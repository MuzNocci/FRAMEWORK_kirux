### KYRUX FRAMEWORK
Criado e desenvolvido por Müller Nocciolli.
Contato: muller.nocciolli@gmail.com



## VISÃO GERAL:
Framework web em Go baseado em SSR, EventBus e Realtime invisível.


## PRINCÍPIOS:
- SSR-first
- Event-driven architecture
- Realtime automático
- Security centralizada
- Dev foca apenas na lógica


## CORE MODULES:

# Bootstrap:
Inicializa todo o framework (env, db, cache, router, realtime)

# Environment:
Leitura de .env e variáveis do sistema

# Settings:
Configuração tipada global

# Router:
Mapeamento de URLs para views

# Render:
Renderização SSR com injeção automática de realtime

# Security:
Autenticação, sessão, criptografia e middleware

# EventBus:
Sistema interno de eventos desacoplados

# Realtime:
WebSocket invisível + bridge com EventBus


## FLUXO DO SISTEMA:
Request -> Middleware -> Router -> View -> Service -> DB
-> EventBus -> Realtime -> Client update automático


## CONCEITO CHAVE:
Kyrux representa execução no "momento certo" (Kairos),
onde eventos dirigem a atualização do sistema em tempo real.