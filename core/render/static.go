package render

import (
	"net/http"
	"os"
	"path/filepath"
)

type multiStatic struct {
	dirs []string
}

func (m *multiStatic) Open(name string) (http.File, error) {
	for _, dir := range m.dirs {
		f, err := http.Dir(dir).Open(name)
		if err == nil {
			return f, nil
		}
	}
	return nil, os.ErrNotExist
}

// collectStaticDirs retorna statics/ da raiz + statics/ de cada app.
func collectStaticDirs(appsDir string) []string {
	dirs := []string{}

	if _, err := os.Stat("statics"); err == nil {
		dirs = append(dirs, "statics")
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return dirs
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(appsDir, e.Name(), "statics")
		if _, err := os.Stat(path); err == nil {
			dirs = append(dirs, path)
		}
	}

	return dirs
}

func StaticHandler(dir string) http.Handler {
	return http.FileServer(http.Dir(dir))
}

func MultiStaticHandler(appsDir string) http.Handler {
	return http.FileServer(&multiStatic{dirs: collectStaticDirs(appsDir)})
}
