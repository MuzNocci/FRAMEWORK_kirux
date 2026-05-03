package database

import (
	"database/sql"
	"fmt"
	"sync"
)

// DB encapsula *sql.DB com o nome do driver.
type DB struct {
	*sql.DB
	Driver string
}

func (db *DB) Transaction(fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Manager gerencia múltiplas conexões SQL nomeadas.
type Manager struct {
	mu    sync.RWMutex
	conns map[string]*DB
}

func NewManager() *Manager {
	return &Manager{conns: make(map[string]*DB)}
}

// Add abre e registra uma conexão nomeada.
// Use "default" como name para a conexão principal.
func (m *Manager) Add(name, driver, dsn string) error {
	db, err := open(driver, dsn)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.conns[name] = db
	m.mu.Unlock()
	return nil
}

// Use retorna a conexão pelo nome (padrão: "default").
func (m *Manager) Use(name ...string) *DB {
	key := "default"
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conns[key]
}

// Close encerra todas as conexões registradas.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, db := range m.conns {
		db.Close()
	}
}

func open(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open [%s]: %w", driver, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping [%s]: %w", driver, err)
	}
	return &DB{DB: db, Driver: driver}, nil
}
