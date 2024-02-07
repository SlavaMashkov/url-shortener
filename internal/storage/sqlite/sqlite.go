package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"log/slog"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	statement, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS url (
			id INTEGER PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL UNIQUE
		);
		CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	_, err = statement.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const fn = "storage.sqlite.SaveURL"

	statement, err := s.db.Prepare("INSERT INTO url (url, alias) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	result, err := statement.Exec(urlToSave, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: execute statement: %w", fn, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURLByAlias(alias string) (string, error) {
	const fn = "storage.sqlite.GetURLByAlias"

	statement, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	var url string

	err = statement.QueryRow(alias).Scan(&url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	return url, nil
}

func (s *Storage) DeleteURLByAlias(alias string) (int64, error) {
	const fn = "storage.sqlite.DeleteURLByAlias"

	statement, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	result, err := statement.Exec(alias)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: rows affected: %w", fn, err)
	}

	if rowsAffected == 0 {
		return 0, storage.ErrURLNotFound
	}

	return rowsAffected, nil
}

func (s *Storage) IsAliasExists(alias string) (bool, error) {
	const fn = "storage.sqlite.IsAliasExist"

	url, err := s.GetURLByAlias(alias)
	if err != nil {
		slog.Info(fn, err)

		if errors.Is(err, storage.ErrURLNotFound) {
			return false, nil
		}

		return true, fmt.Errorf("%s: %w", fn, err)
	}

	if url != "" {
		return true, nil
	}

	return false, nil
}
