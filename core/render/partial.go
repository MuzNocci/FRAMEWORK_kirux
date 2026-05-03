package render

// Partial renderiza um fragmento de template para string.
// Usado para enviar HTML via WebSocket com fw.Realtime.Replace/Append/etc.
func Partial(appName, tmplName string, data map[string]any) (string, error) {
	r := For(appName)

	merged := mergedPool.Get().(map[string]any)
	for k := range merged {
		delete(merged, k)
	}
	for k, v := range data {
		merged[k] = v
	}

	result, err := r.engine.RenderToString(tmplName, merged)
	mergedPool.Put(merged)
	return result, err
}
