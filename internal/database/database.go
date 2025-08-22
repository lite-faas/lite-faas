package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Function struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Code        string    `json:"code"`
	Port        *int      `json:"port"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ContainerID *string   `json:"container_id"`
}

type DB struct {
	db *sql.DB
}

func New(dbPath string) (*DB, error) {
	// Créer le répertoire parent s'il n'existe pas
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS functions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		language TEXT NOT NULL,
		code TEXT NOT NULL,
		port INTEGER UNIQUE,
		status TEXT DEFAULT 'stopped',
		container_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_functions_name ON functions(name);
	CREATE INDEX IF NOT EXISTS idx_functions_port ON functions(port);
	CREATE INDEX IF NOT EXISTS idx_functions_status ON functions(status);
	`

	_, err := db.Exec(query)
	return err
}

func (d *DB) CreateFunction(name, language, code string) (*Function, error) {
	query := `
	INSERT INTO functions (name, language, code, status)
	VALUES (?, ?, ?, 'stopped')
	`

	result, err := d.db.Exec(query, name, language, code)
	if err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetFunctionByID(id)
}

func (d *DB) GetFunctionByID(id int64) (*Function, error) {
	query := `
	SELECT id, name, language, code, port, status, container_id, created_at, updated_at
	FROM functions WHERE id = ?
	`

	var f Function
	err := d.db.QueryRow(query, id).Scan(
		&f.ID, &f.Name, &f.Language, &f.Code, &f.Port, &f.Status, &f.ContainerID,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	return &f, nil
}

func (d *DB) GetFunctionByName(name string) (*Function, error) {
	query := `
	SELECT id, name, language, code, port, status, container_id, created_at, updated_at
	FROM functions WHERE name = ?
	`

	var f Function
	err := d.db.QueryRow(query, name).Scan(
		&f.ID, &f.Name, &f.Language, &f.Code, &f.Port, &f.Status, &f.ContainerID,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	return &f, nil
}

func (d *DB) ListFunctions() ([]*Function, error) {
	query := `
	SELECT id, name, language, code, port, status, container_id, created_at, updated_at
	FROM functions ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []*Function
	for rows.Next() {
		var f Function
		err := rows.Scan(
			&f.ID, &f.Name, &f.Language, &f.Code, &f.Port, &f.Status, &f.ContainerID,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}
		functions = append(functions, &f)
	}

	return functions, nil
}

func (d *DB) UpdateFunction(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE functions SET "
	args := []interface{}{}
	placeholders := []string{}

	for key, value := range updates {
		placeholders = append(placeholders, key+" = ?")
		args = append(args, value)
	}
	placeholders = append(placeholders, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query += fmt.Sprintf("%s WHERE id = ?", strings.Join(placeholders[:len(placeholders)-1], ", "))

	_, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}

	return nil
}

func (d *DB) DeleteFunction(id int64) error {
	query := "DELETE FROM functions WHERE id = ?"
	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	return nil
}

func (d *DB) GetAvailablePort() (int, error) {
	query := `
	SELECT port FROM functions
	WHERE port IS NOT NULL
	ORDER BY port ASC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to get used ports: %w", err)
	}
	defer rows.Close()

	usedPorts := make(map[int]bool)
	for rows.Next() {
		var port int
		if err := rows.Scan(&port); err != nil {
			return 0, fmt.Errorf("failed to scan port: %w", err)
		}
		usedPorts[port] = true
	}

	for port := 9000; port <= 9999; port++ {
		if !usedPorts[port] {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports")
}

func (d *DB) GetFunctionByPort(port int) (*Function, error) {
	query := `
	SELECT id, name, language, code, port, status, container_id, created_at, updated_at
	FROM functions WHERE port = ?
	`

	var f Function
	err := d.db.QueryRow(query, port).Scan(
		&f.ID, &f.Name, &f.Language, &f.Code, &f.Port, &f.Status, &f.ContainerID,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get function by port: %w", err)
	}

	return &f, nil
}
