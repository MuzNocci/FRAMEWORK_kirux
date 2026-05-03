package core

import (
	"kyrux/core/environment"
	"runtime"
	"strconv"
)

type Settings struct {
	InstalledApps []string
	App           AppSettings
	Server        ServerSettings
	Database      DatabaseSettings
	Cache         CacheSettings
	Security      SecuritySettings
}

type AppSettings struct {
	Name    string
	Version string
	Debug   bool
	Env     string
}

type ServerSettings struct {
	Host    string
	Port    string
	Workers int
}

type DatabaseSettings struct {
	Driver string
	DSN    string
}

type CacheSettings struct {
	Driver string
	Addr   string
}

type SecuritySettings struct {
	SecretKey   string
	SessionTTL  int
	AllowedHost []string
}

func intOr(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return fallback
}

func LoadSettings() *Settings {
	return &Settings{
		InstalledApps: []string{},
		App: AppSettings{
			Name:    environment.GetOr("APP_NAME", "kyrux"),
			Version: environment.GetOr("APP_VERSION", "0.1.0"),
			Env:     environment.GetOr("APP_ENV", "production"),
			Debug:   environment.GetOr("APP_ENV", "production") == "development",
		},
		Server: ServerSettings{
			Host:    environment.GetOr("SERVER_HOST", "0.0.0.0"),
			Port:    environment.GetOr("SERVER_PORT", "8000"),
			Workers: intOr(environment.Get("SERVER_WORKERS"), runtime.NumCPU()),
		},
		Database: DatabaseSettings{
			Driver: environment.GetOr("DB_DRIVER", "postgres"),
			DSN:    environment.Get("DB_DSN"),
		},
		Cache: CacheSettings{
			Driver: environment.Get("CACHE_DRIVER"),
			Addr:   environment.Get("CACHE_ADDR"),
		},
		Security: SecuritySettings{
			SecretKey:  environment.GetOr("SECRET_KEY", "change-me"),
			SessionTTL: 3600,
		},
	}
}
