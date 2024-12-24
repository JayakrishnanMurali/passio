package cmd

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newGetCmd(app *app.App) *cobra.Command {
	var (
		copyToClipboard bool
		showPassword    bool
		showNotes       bool
	)

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Retrieve a password entry",
		Long: `Retrieve a password entry by name. 
By default, only shows username and URL. Use flags to show additional information.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			name := args[0]

			// Get entry from storage
			entry, err := app.Storage.GetEntry(name)
			if err != nil {
				return fmt.Errorf("failed to get entry: %w", err)
			}

			var password string
			if showPassword || copyToClipboard {
				password, err = app.DecryptPassword(entry.Password)
				if err != nil {
					return fmt.Errorf("failed to decrypt password: %w", err)
				}
			}

			if copyToClipboard {
				if err := clipboard.WriteAll(password); err != nil {
					return fmt.Errorf("failed to copy to clipboard: %w", err)
				}
				fmt.Println("Password copied to clipboard")

				// Clear clipboard after timeout
				if timeout := app.Config.ClipboardTimeout; timeout > 0 {
					go func() {
						time.Sleep(time.Duration(timeout) * time.Second)
						clipboard.WriteAll("")
					}()
				}
			}

			fmt.Printf("Name: %s\n", entry.Name)
			if entry.Username != "" {
				fmt.Printf("Username: %s\n", entry.Username)
			}
			if entry.URL != "" {
				fmt.Printf("URL: %s\n", entry.URL)
			}
			if showPassword {
				fmt.Printf("Password: %s\n", password)
			}
			if showNotes && entry.Notes != "" {
				fmt.Printf("Notes: %s\n", entry.Notes)
			}
			if len(entry.Tags) > 0 {
				fmt.Printf("Tags: %s\n", entry.Tags)
			}
			fmt.Printf("Created: %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Last modified: %s\n", entry.UpdatedAt.Format("2006-01-02 15:04:05"))

			return nil
		},
	}

	cmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy password to clipboard")
	cmd.Flags().BoolVarP(&showPassword, "show-password", "p", false, "Show password in output")
	cmd.Flags().BoolVarP(&showNotes, "show-notes", "n", false, "Show notes in output")

	return cmd
}
