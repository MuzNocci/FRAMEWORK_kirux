package render

import "runtime"

// RegisterAppFuncs registra as variáveis estáticas do framework no FuncMap.
// Disponíveis nos templates sem ponto: {{ AppName }}, {{ Version }}, etc.
func RegisterAppFuncs(name, version, env, addr string) {
	goVer := runtime.Version()
	AddFunc("AppName", func() string { return name })
	AddFunc("Version", func() string { return version })
	AddFunc("Env", func() string { return env })
	AddFunc("Addr", func() string { return addr })
	AddFunc("GoVersion", func() string { return goVer })
}
