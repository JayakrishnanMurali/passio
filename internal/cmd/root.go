package cmd

import (
	"fmt"
	"syscall"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewRootCmd(app *app.App) *cobra.Command {
	var (
		configFile string
		debug      bool
	)

	cmd := &cobra.Command{
		Use:   "pm",
		Short: "Passio is a secure command-line password manager",
		Long: `Passio is a secure command-line password manager that helps you store and manage your passwords
with strong encryption and easy-to-use commands.

Features:
- Secure password storage with AES-256 encryption
- Password generation with customizable options
- Password security auditing
- Import/export functionality
- Automatic clipboard clearing
- Tags and search functionality`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "init" {
				return nil
			}

			if !app.IsInitialized() {
				return fmt.Errorf("passio is not initialized. Run 'pm init' first")
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.passio/config.json)")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")

	cmd.AddCommand(
		newInitCmd(app),
		newAddCmd(app),
		newGetCmd(app),
		newListCmd(app),
		newUpdateCmd(app),
		newDeleteCmd(app),
		newSearchCmd(app),
		newGenerateCmd(),
		newAuditCmd(app),
		newLockCmd(app),
		newUnlockCmd(app),
		newExportCmd(app),
		newStatsCmd(app),
		newImportCmd(app),
		newConfigCmd(app),
		newBackupCmd(app),
		newRestoreCmd(app),
		newVersionCmd(),
	)

	return cmd
}

func newLockCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Lock passio",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Lock()
			fmt.Println("Password manager locked")
			return nil
		},
	}
}

func newUnlockCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "unlock",
		Short: "Unlock passio",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Enter master password: ")
			password, err := readPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			if err := app.Unlock(password); err != nil {
				return fmt.Errorf("failed to unlock: %w", err)
			}

			fmt.Println("Password manager unlocked")
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Passio version 1.0.0")
		},
	}
}

func readPassword() (string, error) {
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // Print a newline after the password input
	return string(password), nil
}
