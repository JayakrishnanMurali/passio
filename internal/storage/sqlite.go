package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db   *sql.DB
	mu   sync.RWMutex
	path string
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db, path: dbPath}

	return storage, nil
}

func (s *SQLiteStorage) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			username TEXT,
			password BLOB NOT NULL,
			url TEXT,
			notes TEXT,
			tags TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_name ON entries(name)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_username ON entries(username)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_created_at ON entries(created_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to initialize db: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

func (s *SQLiteStorage) AddEntry(entry *Entry) error {
	if err := ValidateEntry(entry); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tags, err := json.Marshal(entry.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO entries (name, username, password, url, notes, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(query,
		entry.Name,
		entry.Username,
		entry.Password,
		entry.URL,
		entry.Notes,
		string(tags),
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrEntryExists
		}
		return fmt.Errorf("failed to add entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	entry.ID = id

	return err
}

func (s *SQLiteStorage) GetEntry(name string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, username, password, url, notes, tags, created_at, updated_at FROM entries WHERE name = ?`

	var entry Entry
	var tagsJSON string

	err := s.db.QueryRow(query, name).Scan(
		&entry.ID,
		&entry.Name,
		&entry.Username,
		&entry.Password,
		&entry.URL,
		&entry.Notes,
		&tagsJSON,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrEntryNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &entry.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &entry, nil
}

func (s *SQLiteStorage) UpdateEntry(entry *Entry) error {
	if err := ValidateEntry(entry); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tags, err := json.Marshal(entry.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE entries
		SET username = ?, password = ?, url = ?, notes = ?, tags = ?, updated_at = ?
		WHERE name = ?
	`

	result, err := s.db.Exec(query,
		entry.Username,
		entry.Password,
		entry.URL,
		entry.Notes,
		string(tags),
		time.Now(),
		entry.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrEntryNotFound
	}

	return nil
}

func (s *SQLiteStorage) DeleteEntry(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM entries WHERE name = ?`

	result, err := s.db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrEntryNotFound
	}

	return nil
}

func (s *SQLiteStorage) ListEntries() ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, username, password, url, notes, tags, created_at, updated_at
			 FROM entries ORDER BY name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		var entry Entry
		var tagsJSON string

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.Username,
			&entry.Password,
			&entry.URL,
			&entry.Notes,
			&tagsJSON,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &entry.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		entries = append(entries, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

func (s *SQLiteStorage) SearchEntries(query string) ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sqlQuery := `
		SELECT id, name, username, password, url, notes, tags, created_at, updated_at
		FROM entries
		WHERE name LIKE ? OR username LIKE ? OR url LIKE ? OR notes LIKE ?
		ORDER BY name
	`

	searchPattern := "%" + query + "%"

	rows, err := s.db.Query(sqlQuery, searchPattern, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search entries: %w", err)
	}

	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		var entry Entry
		var tagsJSON string

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.Username,
			&entry.Password,
			&entry.URL,
			&entry.Notes,
			&tagsJSON,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &entry.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (s *SQLiteStorage) GetEntriesByTag(tag string) ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, name, username, password, url, notes, tags, created_at, updated_at
		FROM entries
		WHERE tags LIKE ?
		ORDER BY name
	`

	searchPattern := "%\"" + tag + "\"%"
	rows, err := s.db.Query(query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries by tag: %w", err)
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		var entry Entry
		var tagsJSON string

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.Username,
			&entry.Password,
			&entry.URL,
			&entry.Notes,
			&tagsJSON,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &entry.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (s *SQLiteStorage) GetStats() (*StorageStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &StorageStats{}

	totalCountQuery := `SELECT COUNT(*) FROM entries`
	err := s.db.QueryRow(totalCountQuery).Scan(&stats.TotalEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to get total entries: %w", err)
	}

	oldestAndNewestQuery := `SELECT MIN(created_at), MAX(created_at) FROM entries`
	err = s.db.QueryRow(oldestAndNewestQuery).Scan(&stats.OldestEntry, &stats.NewestEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to get oldest and newest entries: %w", err)
	}

	passwordAgeQuery := `SELECT updated_at FROM entries`
	var totalAge float64
	rows, err := s.db.Query(passwordAgeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get password age: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var updatedAt time.Time
		if err := rows.Scan(&updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan updated_at: %w", err)
		}
		totalAge += time.Since(updatedAt).Hours() / 24
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating updated_at: %w", err)
	}

	stats.AveragePassAge = totalAge / float64(stats.TotalEntries)

	return stats, nil
}

func (s *SQLiteStorage) Backup(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer backup.Rollback()

	query := `VACUUM INTO ?`
	_, err = backup.Exec(query, path)
	if err != nil {
		return fmt.Errorf("failed to backup database: %w", err)
	}

	return backup.Commit()
}

func (s *SQLiteStorage) Restore(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close current database: %w", err)
	}

	if err := copyFile(path, s.path); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return fmt.Errorf("failed to open restored database: %w", err)
	}

	s.db = db

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
