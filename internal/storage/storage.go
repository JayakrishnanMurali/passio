package storage

import (
	"errors"
	"time"
)

var (
	ErrEntryNotFound      = errors.New("entry not found")
	ErrEntryExists        = errors.New("entry already exists")
	ErrInvalidEntry       = errors.New("invalid entry")
	ErrStorageNotInit     = errors.New("storage not initialized")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrEntryNameIsReq     = errors.New("entry name is required")
	ErrEntryPasswordIsReq = errors.New("entry password is required")
)

type Entry struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  []byte    `json:"password"` // Encrypted password
	URL       string    `json:"url"`
	Notes     string    `json:"notes"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Storage interface {
	// Initialize and cleanup
	Initialize() error
	Close() error

	// CRUD
	AddEntry(entry *Entry) error
	GetEntry(name string) (*Entry, error)
	UpdateEntry(entry *Entry) error
	DeleteEntry(name string) error

	// Query
	ListEntries() ([]*Entry, error)
	SearchEntries(query string) ([]*Entry, error)
	GetEntriesByTag(tag string) ([]*Entry, error)

	// Backup and restore
	Backup(path string) error
	Restore(path string) error

	// Stats
	GetStats() (*StorageStats, error)
}

type StorageStats struct {
	TotalEntries     int       `json:"total_entries"`
	OldestEntry      time.Time `json:"oldest_entry"`
	NewestEntry      time.Time `json:"newest_entry"`
	AveragePassAge   float64   `json:"average_pass_age"` // in days
	WeakPasswords    int       `json:"weak_passwords"`
	ExpiredPasswords int       `json:"expired_passwords"`
}

type SearchOptions struct {
	Query     string   `json:"query"`
	Tags      []string `json:"tags"`
	DateFrom  string   `json:"date_from"`
	DateTo    string   `json:"date_to"`
	SortBy    string   `json:"sort_by"`
	SortOrder string   `json:"sort_order"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
}

func ValidateEntry(entry *Entry) error {
	if entry == nil {
		return ErrInvalidEntry
	}

	if entry.Name == "" {
		return ErrEntryNameIsReq
	}

	if len(entry.Password) == 0 {
		return ErrEntryPasswordIsReq
	}

	return nil
}

func NewEntry(name, username string, password []byte) *Entry {
	now := time.Now()
	return &Entry{
		Name:      name,
		Username:  username,
		Password:  password,
		Tags:      make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type StorageType string

const (
	SQLite StorageType = "sqlite"
)

func NewStorage(storageType string, path string) (Storage, error) {
	switch StorageType(storageType) {
	case SQLite:
		return NewSQLiteStorage(path)
	default:
		return nil, errors.New("unsupported storage type")
	}
}
