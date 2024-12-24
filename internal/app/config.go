package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultConfigDir  = ".passio"
	defaultConfigFile = "config.json"
	defaultDBFile     = "passio.db"
)

type Config struct {
	// Password hash and salt
	MasterHash []byte `json:"master_hash"`
	Salt       []byte `json:"salt"`

	// Storage
	StorageType string `json:"storage_type"`
	DBPath      string `json:"db_path"`

	// App settings
	ConfigPath    string `json:"config_path"`
	LastBackup    string `json:"last_backup"`
	BackupEnabled bool   `json:"backup_enabled"`

	// Security settings
	PasswordLength        int  `json:"password_length"`
	UseSpecialChars       bool `json:"use_special_chars"`
	ClipboardTimeout      int  `json:"clipboard_timeout"`
	AutoLockTimeout       int  `json:"auto_lock_timeout"`
	RequireMasterPassword bool `json:"require_master_password"`
	BackupEncrypted       bool `json:"backup_encrypted"`
	PasswordExpiration    int  `json:"password_expiration"`
}

func loadConfig() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, defaultConfigFile)
	dbPath := filepath.Join(configDir, defaultDBFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := &Config{
			StorageType:           "sqlite",
			DBPath:                dbPath,
			ConfigPath:            configPath,
			PasswordLength:        16,
			UseSpecialChars:       true,
			ClipboardTimeout:      30,
			AutoLockTimeout:       300,
			RequireMasterPassword: true,
			BackupEncrypted:       true,
			PasswordExpiration:    90,
		}

		return config, config.Save()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.DBPath == "" {
		config.DBPath = dbPath
	}

	if config.ConfigPath == "" {
		config.ConfigPath = configPath
	}

	return &config, nil
}

func (c *Config) Save() error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(c.ConfigPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config to file with restricted permissions
	if err := os.WriteFile(c.ConfigPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, defaultConfigDir)
	return configDir, nil
}

func (c *Config) SetMasterKey(masterKey, salt []byte) error {
	c.MasterHash = masterKey
	c.Salt = salt
	return c.Save()
}

func (c *Config) ValidateMasterPassword(app *App, password string) bool {
	derivedKey := app.Encryption.DeriveKey(password, c.Salt)
	return string(derivedKey) == string(c.MasterHash)
}

func (c *Config) GetConfigValue(key string) interface{} {
	switch key {
	case "password_length":
		return c.PasswordLength
	case "use_special_chars":
		return c.UseSpecialChars
	case "clipboard_timeout":
		return c.ClipboardTimeout
	case "auto_lock_timeout":
		return c.AutoLockTimeout
	case "require_master_pass":
		return c.RequireMasterPassword
	case "backup_encrypted":
		return c.BackupEncrypted
	case "password_expiration":
		return c.PasswordExpiration
	default:
		return nil
	}
}

func (c *Config) SetConfigValue(key string, value interface{}) error {
	switch key {
	case "password_length":
		if v, ok := value.(int); ok {
			c.PasswordLength = v
		} else {
			return fmt.Errorf("invalid value type for password_length")
		}
	case "use_special_chars":
		if v, ok := value.(bool); ok {
			c.UseSpecialChars = v
		} else {
			return fmt.Errorf("invalid value type for use_special_chars")
		}
	case "clipboard_timeout":
		if v, ok := value.(int); ok {
			c.ClipboardTimeout = v
		} else {
			return fmt.Errorf("invalid value type for clipboard_timeout")
		}
	case "auto_lock_timeout":
		if v, ok := value.(int); ok {
			c.AutoLockTimeout = v
		} else {
			return fmt.Errorf("invalid value type for auto_lock_timeout")
		}
	case "require_master_pass":
		if v, ok := value.(bool); ok {
			c.RequireMasterPassword = v
		} else {
			return fmt.Errorf("invalid value type for require_master_pass")
		}
	case "backup_encrypted":
		if v, ok := value.(bool); ok {
			c.BackupEncrypted = v
		} else {
			return fmt.Errorf("invalid value type for backup_encrypted")
		}
	case "password_expiration":
		if v, ok := value.(int); ok {
			c.PasswordExpiration = v
		} else {
			return fmt.Errorf("invalid value type for password_expiration")
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return c.Save()
}
