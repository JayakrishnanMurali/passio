package cmd

import (
	"fmt"
	"strings"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newUpdateCmd(app *app.App) *cobra.Command {
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
		Use:   "update <name>",
		Short: "Update an existing password entry",
		Long: `Update an existing password entry in the password manager.
Only specified fields will be updated. Use --generate to create a new password.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			name := args[0]

			// Get existing entry
			entry, err := app.Storage.GetEntry(name)
			if err != nil {
				return fmt.Errorf("failed to get entry: %w", err)
			}

			// Update fields if provided
			if username != "" {
				entry.Username = username
			}

			if generate || password != "" {
				var newPassword string
				if generate {
					var err error
					newPassword, err = generatePassword(length, special)
					if err != nil {
						return fmt.Errorf("failed to generate password: %w", err)
					}
					fmt.Printf("Generated new password: %s\n", newPassword)
				} else {
					newPassword = password
				}

				// Encrypt the new password
				encryptedPass, err := app.EncryptPassword(newPassword)
				if err != nil {
					return fmt.Errorf("failed to encrypt password: %w", err)
				}
				entry.Password = encryptedPass
			}

			if url != "" {
				entry.URL = url
			}

			if notes != "" {
				entry.Notes = notes
			}

			if tags != "" {
				tagList := strings.Split(tags, ",")
				for i, tag := range tagList {
					tagList[i] = strings.TrimSpace(tag)
				}
				entry.Tags = tagList
			}

			// Update entry in storage
			if err := app.Storage.UpdateEntry(entry); err != nil {
				return fmt.Errorf("failed to update entry: %w", err)
			}

			fmt.Printf("Successfully updated entry: %s\n", name)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&username, "username", "u", "", "New username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "New password")
	cmd.Flags().StringVar(&url, "url", "", "New URL")
	cmd.Flags().StringVar(&notes, "notes", "", "New notes")
	cmd.Flags().StringVar(&tags, "tags", "", "New comma-separated list of tags")
	cmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate a new password")
	cmd.Flags().IntVarP(&length, "length", "l", 16, "Length of generated password")
	cmd.Flags().BoolVarP(&special, "special", "s", true, "Include special characters in generated password")

	return cmd
}
