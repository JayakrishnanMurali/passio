package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newBackupCmd(app *app.App) *cobra.Command {
	var (
		outputDir string
		compress  bool
	)

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of the password database",
		Long: `Create a backup of the password database.
Backups are encrypted by default and can be compressed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			// Create backup directory if it doesn't exist
			if outputDir == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}
				outputDir = filepath.Join(homeDir, ".pm", "backups")
			}

			if err := os.MkdirAll(outputDir, 0700); err != nil {
				return fmt.Errorf("failed to create backup directory: %w", err)
			}

			// Generate backup filename
			timestamp := time.Now().Format("20060102_150405")
			filename := fmt.Sprintf("pm_backup_%s.db", timestamp)
			if compress {
				filename += ".gz"
			}
			backupPath := filepath.Join(outputDir, filename)

			// Create backup
			if err := app.Storage.Backup(backupPath); err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}

			fmt.Printf("Successfully created backup: %s\n", backupPath)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for backup")
	cmd.Flags().BoolVarP(&compress, "compress", "c", false, "Compress the backup file")

	return cmd
}
