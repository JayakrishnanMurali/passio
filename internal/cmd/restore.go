package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newRestoreCmd(app *app.App) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "restore <backup-file>",
		Short: "Restore from a backup file",
		Long: `Restore the password database from a backup file.
This will replace the current database with the backup.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			backupFile := args[0]
			if _, err := os.Stat(backupFile); err != nil {
				return fmt.Errorf("backup file not found: %w", err)
			}

			// Confirm restore unless force flag is set
			if !force {
				fmt.Print("WARNING: This will replace your current database. Continue? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Restore cancelled")
					return nil
				}
			}

			// Perform restore
			if err := app.Storage.Restore(backupFile); err != nil {
				return fmt.Errorf("restore failed: %w", err)
			}

			fmt.Println("Successfully restored from backup")
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
