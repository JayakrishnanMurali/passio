package cmd

import (
	"fmt"
	"strings"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newDeleteCmd(app *app.App) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a password entry",
		Long: `Delete a password entry by name. 
Use --force to skip confirmation prompt.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			name := args[0]

			// Confirm deletion unless force flag is set
			if !force {
				fmt.Printf("Are you sure you want to delete entry '%s'? [y/N]: ", name)
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Deletion cancelled")
					return nil
				}
			}

			if err := app.Storage.DeleteEntry(name); err != nil {
				return fmt.Errorf("failed to delete entry: %w", err)
			}

			fmt.Printf("Successfully deleted entry: %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
