package app

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/crypto"
	"github.com/jayakrishnanMurali/passio/internal/storage"
)

type App struct {
	Storage    storage.Storage
	Encryption crypto.Encryption
	Config     *Config

	// Session
	isLocked     bool
	lastActivity time.Time
	mu           sync.RWMutex
}

func New() (*App, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	storage, err := storage.NewStorage(config.StorageType, config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	encryptions := crypto.NewAESEncryption()

	app := &App{
		Storage:      storage,
		Encryption:   encryptions,
		Config:       config,
		isLocked:     true,
		lastActivity: time.Now(),
	}

	return app, nil
}

func (a *App) IsInitialized() bool {
	return len(a.Config.MasterHash) > 0
}

func (a *App) Lock() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.isLocked = true
}

func (a *App) Unlock(masterPassword string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.Config.ValidateMasterPassword(a, masterPassword) {
		return errors.New("invalid master password")
	}

	a.isLocked = false
	a.lastActivity = time.Now()
	return nil
}

func (a *App) IsLocked() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.isLocked
}

func (a *App) UpdateActivity() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastActivity = time.Now()
}

func (a *App) CheckAutoLock() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isLocked && a.Config.AutoLockTimeout > 0 {
		inactiveTime := time.Since(a.lastActivity)
		if inactiveTime.Seconds() >= float64(a.Config.AutoLockTimeout) {
			a.isLocked = true
		}
	}
}

func (a *App) DecryptMasterPassword(encryptedPassword []byte) (string, error) {
	if a.isLocked {
		return "", errors.New("passio is locked")
	}

	decrypted, err := a.Encryption.Decrypt(encryptedPassword, a.Config.MasterHash)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt master password: %w", err)
	}

	return string(decrypted), nil
}

func (a *App) EncryptPassword(password string) ([]byte, error) {
	if a.IsLocked() {
		return nil, errors.New("password manager is locked")
	}

	encrypted, err := a.Encryption.Encrypt([]byte(password), a.Config.MasterHash)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	return encrypted, nil
}

func (a *App) Close() error {
	if err := a.Storage.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}
	return nil
}

func (a *App) CheckPasswordHealth(password string) map[string]bool {
	return map[string]bool{
		"length":       len(password) >= a.Config.PasswordLength,
		"uppercase":    containsUppercase(password),
		"lowercase":    containsLowercase(password),
		"numbers":      containsNumbers(password),
		"specialChars": containsSpecialChars(password),
		"notCommon":    !isCommonPassword(password),
	}
}

func containsUppercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func containsNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialChars(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, c := range s {
		for _, sc := range specialChars {
			if c == sc {
				return true
			}
		}
	}
	return false
}

func isCommonPassword(password string) bool {
	// TODO: Check against a list of common passwords from a file
	commonPasswords := map[string]bool{
		"password": true,
		"123456":   true,
		"qwerty":   true,
	}
	return commonPasswords[password]
}
