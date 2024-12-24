package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newConfigCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long: `Manage configuration settings for the password manager.
Use 'get' to view settings and 'set' to modify them.`,
	}

	cmd.AddCommand(newConfigGetCmd(app))
	cmd.AddCommand(newConfigSetCmd(app))

	return cmd
}

func newConfigGetCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get [setting]",
		Short: "Get configuration settings",
		Long: `Get the current value of a configuration setting.
If no setting is specified, all settings are displayed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Display all settings
				fmt.Println("Current configuration:")
				fmt.Printf("password_length: %d\n", app.Config.PasswordLength)
				fmt.Printf("use_special_chars: %v\n", app.Config.UseSpecialChars)
				fmt.Printf("clipboard_timeout: %d seconds\n", app.Config.ClipboardTimeout)
				fmt.Printf("auto_lock_timeout: %d seconds\n", app.Config.AutoLockTimeout)
				fmt.Printf("require_master_pass: %v\n", app.Config.RequireMasterPassword)
				fmt.Printf("backup_encrypted: %v\n", app.Config.BackupEncrypted)
				fmt.Printf("password_expiration: %d days\n", app.Config.PasswordExpiration)
				return nil
			}

			// Get specific setting
			setting := args[0]
			value := app.Config.GetConfigValue(setting)
			if value == nil {
				return fmt.Errorf("unknown setting: %s", setting)
			}

			fmt.Printf("%s: %v\n", setting, value)
			return nil
		},
	}
}

func newConfigSetCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "set <setting> <value>",
		Short: "Set configuration settings",
		Long: `Set the value of a configuration setting.
Available settings:
  - password_length: Minimum length for generated passwords (int)
  - use_special_chars: Whether to use special characters in generated passwords (bool)
  - clipboard_timeout: Time in seconds before clipboard is cleared (int)
  - auto_lock_timeout: Time in seconds of inactivity before auto-lock (int)
  - require_master_pass: Whether to require master password for sensitive operations (bool)
  - backup_encrypted: Whether to encrypt backup files (bool)
  - password_expiration: Number of days before passwords are considered expired (int)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			setting := args[0]
			valueStr := args[1]

			var value interface{}
			var err error

			// Parse value based on setting type
			switch setting {
			case "password_length", "clipboard_timeout", "auto_lock_timeout", "password_expiration":
				value, err = strconv.Atoi(valueStr)
				if err != nil {
					return fmt.Errorf("invalid integer value: %s", valueStr)
				}
			case "use_special_chars", "require_master_pass", "backup_encrypted":
				valueLower := strings.ToLower(valueStr)
				if valueLower == "true" || valueLower == "1" || valueLower == "yes" {
					value = true
				} else if valueLower == "false" || valueLower == "0" || valueLower == "no" {
					value = false
				} else {
					return fmt.Errorf("invalid boolean value: %s", valueStr)
				}
			default:
				return fmt.Errorf("unknown setting: %s", setting)
			}

			// Validate values
			switch setting {
			case "password_length":
				if v := value.(int); v < 8 {
					return fmt.Errorf("password length must be at least 8")
				}
			case "clipboard_timeout", "auto_lock_timeout":
				if v := value.(int); v < 0 {
					return fmt.Errorf("timeout values must be non-negative")
				}
			case "password_expiration":
				if v := value.(int); v < 0 {
					return fmt.Errorf("expiration days must be non-negative")
				}
			}

			// Update configuration
			if err := app.Config.SetConfigValue(setting, value); err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}

			fmt.Printf("Successfully updated %s to %v\n", setting, value)
			return nil
		},
	}
}
