package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
	mu   sync.RWMutex
}

func Open(dbPath string) (*DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", dbPath)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		last_login TEXT
	)`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("create users table: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) RLock()   { d.mu.RLock() }
func (d *DB) RUnlock() { d.mu.RUnlock() }
func (d *DB) Lock()    { d.mu.Lock() }
func (d *DB) Unlock()  { d.mu.Unlock() }

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn.Query(query, args...)
}

func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn.QueryRow(query, args...)
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.conn.Exec(query, args...)
}

func (d *DB) Begin() (*sql.Tx, error) {
	d.mu.Lock()
	return d.conn.Begin()
}

func (d *DB) UnlockAfterTx() {
	d.mu.Unlock()
}
