package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"

	"time"
	"url-shortener1/internal/model"
)

var (
	ErrUnique        = errors.New("short URL already exists")
	ErrNotFound      = errors.New("short URL not found")
	ErrNoRowsUpdated = errors.New("no rows updated for short URL")
)

type Storage struct {
	db *sql.DB
}

func Init() (*Storage, error) {
	const op = "storage.sqlite.Init"
	db, err := sql.Open("sqlite3", "./url-shortener.db")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %w", op, err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %w", op, err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS urls (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_url TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL, 
		visits INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	);
	CREATE INDEX IF NOT EXISTS urls_short_url_idx ON urls(short_url);
	`

	_, err = db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create table: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Storage) Save(sh *model.Shorten) error {
	const op = "storage.sqlite.Save"

	if sh == nil {
		return fmt.Errorf("%s: shorten model is nil", op)
	}

	now := time.Now()
	sh.CreatedAt = now
	sh.UpdatedAt = now

	stmt, err := s.db.Prepare("INSERT INTO urls(short_url, original_url, visits, created_at, updated_at) VALUES (?,?,?,?,?)")
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(sh.ShortURL, sh.OriginalURL, sh.Visits, now, now)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: short URL already exists: %w", op, err)
		}
		return fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	return nil
}

func (s *Storage) IncVisits(shortURL string) error {
	const op = "storage.sqlite.IncVisits"

	if shortURL == "" {
		return fmt.Errorf("%s: short URL is empty", op)
	}

	stmt, err := s.db.Prepare("UPDATE urls SET visits = visits + 1, updated_at = CURRENT_TIMESTAMP WHERE short_url = ?")
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(shortURL)
	if err != nil {
		return fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows updated for short URL: %s", op, shortURL)
	}

	return nil
}

func (s *Storage) GetStats(shortURL string) (*model.Shorten, error) {
	const op = "storage.sqlite.GetStats"

	if shortURL == "" {
		return nil, fmt.Errorf("%s: short URL is empty", op)
	}

	stmt, err := s.db.Prepare("SELECT id, short_url, original_url, visits, created_at, updated_at FROM urls WHERE short_url = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var sh model.Shorten
	err = stmt.QueryRow(shortURL).Scan(&sh.ID, &sh.ShortURL, &sh.OriginalURL, &sh.Visits, &sh.CreatedAt, &sh.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: short URL not found: %s", op, shortURL)
		}
		return nil, fmt.Errorf("%s: failed to query row: %w", op, err)
	}

	return &sh, nil
}
