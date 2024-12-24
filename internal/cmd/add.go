package cmd

import (
	"fmt"
	"strings"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/jayakrishnanMurali/passio/internal/storage"
	"github.com/spf13/cobra"
)

func newAddCmd(app *app.App) *cobra.Command {
	var (
		username string
		password string
		url      string
		notes    string
		tags     string
		generate bool
		length   int
		special  bool
	)

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new password entry",
		Long: `Add a new password entry to the passio.
If no password is provided, one will be generated using the specified options.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("passio is locked. Please unlock first")
			}

			name := args[0]

			// Generate password if requested or no password provided
			if generate || password == "" {
				var err error
				password, err = generatePassword(length, special)
				if err != nil {
					return fmt.Errorf("failed to generate password: %w", err)
				}
				fmt.Printf("Generated password: %s\n", password)
			}

			// Encrypt the password
			encryptedPass, err := app.EncryptPassword(password)
			if err != nil {
				return fmt.Errorf("failed to encrypt password: %w", err)
			}

			// Parse tags
			tagList := make([]string, 0)
			if tags != "" {
				tagList = strings.Split(tags, ",")
				for i, tag := range tagList {
					tagList[i] = strings.TrimSpace(tag)
				}
			}

			// Create and validate entry
			entry := &storage.Entry{
				Name:     name,
				Username: username,
				Password: encryptedPass,
				URL:      url,
				Notes:    notes,
				Tags:     tagList,
			}

			// Add entry to storage
			if err := app.Storage.AddEntry(entry); err != nil {
				return fmt.Errorf("failed to add entry: %w", err)
			}

			fmt.Printf("Successfully added entry: %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for the entry")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the entry (optional)")
	cmd.Flags().StringVar(&url, "url", "", "URL associated with the entry")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes for the entry")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated list of tags")
	cmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate a password")
	cmd.Flags().IntVarP(&length, "length", "l", 16, "Length of generated password")
	cmd.Flags().BoolVarP(&special, "special", "s", true, "Include special characters in generated password")

	return cmd

}
