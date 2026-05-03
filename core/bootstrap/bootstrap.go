package bootstrap

import (
	"context"
	"fmt"
	"kyrux/core/bootstrap/welcome"
	"kyrux/core/cache"
	"kyrux/core/database"
	kyerrors "kyrux/core/errors"
	"kyrux/core/environment"
	"kyrux/core/events"
	"kyrux/core/hotreload"
	"kyrux/core/realtime"
	"kyrux/core/render"
	"kyrux/core/router"
	"kyrux/core/security/auth"
	"kyrux/core/security/csrf"
	secmiddleware "kyrux/core/security/middleware"
	"kyrux/core/security/session"
	"kyrux/core/settings"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"
)

type Framework struct {
	Settings *settings.Settings
	Router   *router.Router
	Events   *events.Bus
	Realtime *realtime.Hub
	DB       *database.Manager
	Cache    *cache.Cache
	Auth     *auth.Authenticator
	Sessions *session.Store
}

func Init(envPath string) (*Framework, error) {
	if err := environment.Load(envPath); err != nil {
		return nil, fmt.Errorf("bootstrap: env: %w", err)
	}

	cfg := settings.Load()

	render.SetDebug(cfg.App.Debug)
	kyerrors.SetDebug(cfg.App.Debug)
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	render.RegisterAppFuncs(cfg.App.Name, cfg.App.Version, cfg.App.Env, addr)
	csrf.RegisterFuncs()

	bus := events.NewBus()
	hub := realtime.NewHub(bus)
	r := router.New()
	kyerrors.SetRouteListFunc(r.Routes)
	r.Use(secmiddleware.Recovery())
	r.Use(secmiddleware.AllowedHosts(cfg.Security.AllowedHost, cfg.App.Debug))
	r.Use(csrf.Middleware)
	a := auth.New(cfg.Security.SecretKey)
	store := session.NewStore(time.Duration(cfg.Security.SessionTTL) * time.Second)

	dbm := database.NewManager()
	if cfg.Database.Enabled {
		if err := dbm.Add("default", cfg.Database.Driver, cfg.Database.DSN); err != nil {
			return nil, fmt.Errorf("bootstrap: db: %w", err)
		}
		log.Println("bootstrap: database connected")
	} else {
		log.Println("bootstrap: database disabled (DB_ENABLED=false)")
	}

	f := &Framework{
		Settings: cfg,
		Router:   r,
		Events:   bus,
		Realtime: hub,
		DB:       dbm,
		Auth:     a,
		Sessions: store,
	}

	if cfg.Cache.Enabled {
		f.Cache = cache.New()
		log.Println("bootstrap: cache enabled")
	} else {
		log.Println("bootstrap: cache disabled (CACHE_ENABLED=false)")
	}

	for _, appName := range cfg.InstalledApps {
		if fn, ok := registry[appName]; ok {
			fn(r)
			log.Printf("bootstrap: app '%s' registrado\n", appName)
		} else {
			log.Printf("bootstrap: app '%s' listado em InstalledApps mas não importado\n", appName)
		}
	}

	welcome.RegisterIfNeeded(r)

	r.Internal("GET /kyrux/websocket/ws/", func(ctx *router.Context) {
		hub.ServeHTTP(ctx.Writer, ctx.Request)
	})

	static := render.MultiStaticHandler("apps")
	r.HandlePrefix("GET /static/", http.StripPrefix("/static/", static))

	if cfg.App.Debug {
		lr := hotreload.NewHub()
		lr.Watch("apps", "statics")
		r.HandlePrefix("GET /__kyrux_reload__", lr)
		log.Println("bootstrap: hotreload ativo")

		go func() {
			log.Println("bootstrap: pprof disponível em http://localhost:6060/debug/pprof/")
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	return f, nil
}

func (f *Framework) Run() error {
	runtime.GOMAXPROCS(f.Settings.Server.Workers)

	if gc := environment.Get("RUNTIME_GOGC"); gc != "" {
		if n, err := strconv.Atoi(gc); err == nil {
			debug.SetGCPercent(n)
			log.Printf("bootstrap: GOGC=%d\n", n)
		}
	}

	addr := f.Settings.Server.Host + ":" + f.Settings.Server.Port
	fmt.Printf("Kyrux running on http://%s (workers: %d)\n", addr, f.Settings.Server.Workers)

	srv := &http.Server{
		Addr:           addr,
		Handler:        f.Router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		fmt.Printf("\n[%s] cleaning...\n", time.Now().Format("15:04:05"))
		fmt.Printf("signal: %s\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		fmt.Printf("[%s] see you again.\n", time.Now().Format("15:04:05"))
		close(done)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	<-done
	return nil
}
