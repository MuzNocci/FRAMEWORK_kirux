package orm

import (
	"database/sql"
	"fmt"
	"reflect"
)

// scanRows converte sql.Rows em []T mapeando colunas para campos do struct via ModelMeta.
func scanRows[T any](rows *sql.Rows, meta *ModelMeta) ([]T, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("orm: columns: %w", err)
	}

	// Mapa coluna → GoIndex, construído uma vez por query.
	colToField := make(map[string]int, len(meta.Fields))
	for _, f := range meta.Fields {
		colToField[f.Column] = f.GoIndex
	}

	dests := make([]any, len(cols))
	var results []T

	for rows.Next() {
		var zero T
		v := reflect.ValueOf(&zero).Elem()

		for i, col := range cols {
			if idx, ok := colToField[col]; ok {
				fv := v.Field(idx)
				if fv.CanAddr() {
					dests[i] = fv.Addr().Interface()
					continue
				}
			}
			// Coluna sem campo correspondente: descarta silenciosamente.
			dests[i] = new(any)
		}

		if err := rows.Scan(dests...); err != nil {
			return nil, fmt.Errorf("orm: scan: %w", err)
		}
		results = append(results, zero)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("orm: rows: %w", err)
	}
	return results, nil
}
