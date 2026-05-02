package environment

import (
	"os"
	"strings"
)

var loaded = map[string]string{}

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		loaded[key] = value
		os.Setenv(key, value)
	}
	return nil
}

func Get(key string) string {
	if v, ok := loaded[key]; ok {
		return v
	}
	return os.Getenv(key)
}

func GetOr(key, fallback string) string {
	if v := Get(key); v != "" {
		return v
	}
	return fallback
}
