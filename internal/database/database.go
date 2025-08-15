package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Function struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Code        string    `json:"code"`
	Port        int       `json:"port"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Database struct {
	db *sql.DB
}

func New(dbPath string) (*Database, error) {
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

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS functions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		language TEXT NOT NULL,
		code TEXT NOT NULL,
		port INTEGER UNIQUE NOT NULL,
		status TEXT NOT NULL DEFAULT 'stopped',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(query)
	return err
}

func (d *Database) CreateFunction(fn *Function) error {
	query := `
	INSERT INTO functions (name, language, code, port, status)
	VALUES (?, ?, ?, ?, ?)
	`

	result, err := d.db.Exec(query, fn.Name, fn.Language, fn.Code, fn.Port, fn.Status)
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	fn.ID = int(id)
	return nil
}

func (d *Database) GetFunction(name string) (*Function, error) {
	query := `SELECT id, name, language, code, port, status, created_at, updated_at FROM functions WHERE name = ?`

	var fn Function
	err := d.db.QueryRow(query, name).Scan(
		&fn.ID, &fn.Name, &fn.Language, &fn.Code, &fn.Port, &fn.Status,
		&fn.CreatedAt, &fn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	return &fn, nil
}

func (d *Database) ListFunctions() ([]*Function, error) {
	query := `SELECT id, name, language, code, port, status, created_at, updated_at FROM functions ORDER BY name`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query functions: %w", err)
	}
	defer rows.Close()

	var functions []*Function
	for rows.Next() {
		var fn Function
		err := rows.Scan(
			&fn.ID, &fn.Name, &fn.Language, &fn.Code, &fn.Port, &fn.Status,
			&fn.CreatedAt, &fn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}
		functions = append(functions, &fn)
	}

	return functions, nil
}

func (d *Database) UpdateFunction(fn *Function) error {
	query := `
	UPDATE functions
	SET language = ?, code = ?, port = ?, status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE name = ?
	`

	_, err := d.db.Exec(query, fn.Language, fn.Code, fn.Port, fn.Status, fn.Name)
	if err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}

	return nil
}

func (d *Database) DeleteFunction(name string) error {
	query := `DELETE FROM functions WHERE name = ?`

	_, err := d.db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	return nil
}

func (d *Database) UpdateFunctionStatus(name, status string) error {
	query := `UPDATE functions SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE name = ?`

	_, err := d.db.Exec(query, status, name)
	if err != nil {
		return fmt.Errorf("failed to update function status: %w", err)
	}

	return nil
}

func (d *Database) GetAvailablePort() (int, error) {
	query := `SELECT port FROM functions ORDER BY port DESC LIMIT 1`

	var port int
	err := d.db.QueryRow(query).Scan(&port)
	if err == sql.ErrNoRows {
		return 9000, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get last port: %w", err)
	}

	return port + 1, nil
}
