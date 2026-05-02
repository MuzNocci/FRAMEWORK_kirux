package welcome

import (
	"html/template"
	"kyrux/core/environment"
	"kyrux/core/router"
	"runtime"
)

type pageData struct {
	AppName   string
	Version   string
	Env       string
	Addr      string
	GoVersion string
}

func RegisterIfNeeded(r *router.Router) {
	if r.HasRoute("GET /") {
		return
	}
	r.Handle("GET /", handler)
}

func handler(ctx *router.Context) {
	d := pageData{
		AppName:   environment.GetOr("APP_NAME", "kyrux"),
		Version:   environment.GetOr("APP_VERSION", "0.1.0"),
		Env:       environment.GetOr("APP_ENV", "production"),
		Addr:      environment.GetOr("SERVER_HOST", "0.0.0.0") + ":" + environment.GetOr("SERVER_PORT", "8080"),
		GoVersion: runtime.Version(),
	}
	ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = welcomeTpl.Execute(ctx.Writer, d)
}

var welcomeTpl = template.Must(template.New("welcome").Parse(welcomeHTML))

const welcomeHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.AppName}} | Framework Web em Go</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    :root {
      --go-blue:       #00ACD7;
      --go-blue-light: #5DC9E2;
      --go-blue-dark:  #00758D;
      --bg:            #0D1117;
      --surface:       #161B22;
      --border:        #1E2A38;
      --text:          #E6EDF3;
      --muted:         #8B949E;
    }

    html, body { height: 100%; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: var(--bg);
      color: var(--text);
      display: flex;
      flex-direction: column;
      overflow: auto;
    }

    header {
      border-bottom: 1px solid var(--border);
      padding: .9rem 2.5rem;
      display: flex;
      align-items: center;
      gap: .75rem;
      flex-shrink: 0;
    }

    .logo-mark {
      width: 28px; height: 28px;
      background: var(--go-blue);
      border-radius: 6px;
      display: flex; align-items: center; justify-content: center;
      font-weight: 800; font-size: .9rem;
      color: #fff; letter-spacing: -.5px;
    }

    header span { font-size: 1rem; font-weight: 600; color: var(--text); }

    header .version {
      margin-left: auto;
      font-size: .7rem; color: var(--muted);
      background: var(--border);
      padding: .2rem .55rem; border-radius: 999px;
    }

    main {
      flex: 1; display: flex; flex-direction: column;
      align-items: center; justify-content: center;
      padding: 2.5rem 2rem; text-align: center; gap: 1.75rem;
    }

    .status-badge {
      display: inline-flex; align-items: center; gap: .4rem;
      background: rgba(0,172,215,.1);
      border: 1px solid rgba(0,172,215,.3);
      color: var(--go-blue-light);
      font-size: .75rem; font-weight: 500;
      padding: .3rem .8rem; border-radius: 999px; letter-spacing: .02em;
    }

    .status-dot {
      width: 6px; height: 6px; border-radius: 50%;
      background: var(--go-blue);
      animation: pulse 2s ease-in-out infinite;
    }

    @keyframes pulse {
      0%, 100% { opacity: 1; transform: scale(1); }
      50%       { opacity: .4; transform: scale(.8); }
    }

    h1 {
      font-size: clamp(1.6rem, 3.5vw, 2.8rem);
      font-weight: 800; line-height: 1.15; letter-spacing: -.03em;
    }

    h1 .highlight { color: var(--go-blue); }

    .subtitle {
      font-size: clamp(.85rem, 1.5vw, 1rem);
      color: var(--muted); max-width: 480px; line-height: 1.65;
    }

    .cards {
      display: grid; grid-template-columns: repeat(4, 1fr);
      gap: 1rem; width: 100%; max-width: 820px;
    }

    .card {
      background: var(--surface); border: 1px solid var(--border);
      border-radius: 12px; padding: 1.25rem 1.1rem; text-align: left;
      transition: border-color .2s, transform .2s;
    }

    .card:hover { border-color: var(--go-blue-dark); transform: translateY(-2px); }

    .card-icon { font-size: 1.2rem; margin-bottom: .55rem; }
    .card h3 { font-size: .85rem; font-weight: 600; color: var(--text); margin-bottom: .3rem; }
    .card p  { font-size: .75rem; color: var(--muted); line-height: 1.5; }

    .info-grid {
      display: flex; gap: 3rem; flex-wrap: wrap; justify-content: center;
      border-top: 1px solid var(--border); padding-top: 1.5rem;
      width: 100%; max-width: 820px;
    }

    .info-item  { display: flex; flex-direction: column; gap: .2rem; text-align: left; }
    .info-label { font-size: .65rem; color: var(--muted); text-transform: uppercase; letter-spacing: .08em; }
    .info-value { font-size: .85rem; font-weight: 600; color: var(--go-blue-light); font-family: "SF Mono","Fira Code",monospace; }
    .info-value.ok { color: #3FB950; }

    footer {
      border-top: 1px solid var(--border);
      padding: .75rem 2rem; text-align: center;
      font-size: .82rem; color: var(--muted); flex-shrink: 0;
    }

    footer a { color: var(--go-blue); text-decoration: none; }
    footer a:hover { text-decoration: underline; }

    @media (max-width: 640px) {
      body { overflow: auto; }
      main { padding: 2rem 1.25rem; gap: 1.5rem; overflow: visible; }
      .cards { grid-template-columns: repeat(2, 1fr); }
      .info-grid { gap: 1.25rem; }
      h1 { font-size: 2rem; }
      .subtitle { font-size: .9rem; }
      footer { font-size: .9rem; padding: 1.25rem 2rem; }
    }

    @media (max-width: 380px) {
      .cards { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>

  <header>
    <div class="logo-mark">K</div>
    <span>{{.AppName}}</span>
    <span class="version">v{{.Version}}</span>
  </header>

  <main>
    <div class="status-badge">
      <span class="status-dot"></span>
      Servidor operacional
    </div>

    <h1>
      Bem-vindo ao<br>
      <span class="highlight">Kyrux Framework</span>
    </h1>

    <p class="subtitle">
      Framework web em Go baseado em SSR, EventBus e Realtime invisível.
      Seu servidor está instalado e funcionando corretamente.
    </p>

    <div class="cards">
      <div class="card">
        <div class="card-icon">&#x26A1;</div>
        <h3>SSR-First</h3>
        <p>Renderização server-side com injeção automática de realtime.</p>
      </div>
      <div class="card">
        <div class="card-icon">&#x1F4E1;</div>
        <h3>Event-Driven</h3>
        <p>EventBus interno desacoplado para comunicação entre módulos.</p>
      </div>
      <div class="card">
        <div class="card-icon">&#x1F510;</div>
        <h3>Security</h3>
        <p>Auth, sessão, criptografia e middleware centralizados.</p>
      </div>
      <div class="card">
        <div class="card-icon">&#x1F501;</div>
        <h3>Realtime</h3>
        <p>WebSocket invisível com bridge automático ao EventBus.</p>
      </div>
    </div>

    <div class="info-grid">
      <div class="info-item">
        <span class="info-label">Ambiente</span>
        <span class="info-value">{{.Env}}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Endereço</span>
        <span class="info-value">{{.Addr}}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Status</span>
        <span class="info-value ok">&#x25CF; online</span>
      </div>
      <div class="info-item">
        <span class="info-label">Runtime</span>
        <span class="info-value">{{.GoVersion}}</span>
      </div>
    </div>
  </main>

  <footer>
    Kyrux Framework &mdash; construído com <a href="https://go.dev" target="_blank">Go</a>
    &nbsp;&middot;&nbsp;
    desenvolvido por <a href="mailto:muller.nocciolli@gmail.com">M&uuml;ller Nocciolli</a>
  </footer>

</body>
</html>`
