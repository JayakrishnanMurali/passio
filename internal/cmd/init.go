package cmd

import (
	"crypto/rand"
	"fmt"
	"syscall"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newInitCmd(app *app.App) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Passio",
		Long:  "Initialize Passio by creating a new password database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsInitialized() && !force {
				return fmt.Errorf("passio is already initialized. Use --force to reinitialize")
			}

			masterPass, err := getMasterPassword()
			if err != nil {
				return fmt.Errorf("failed to get master password: %w", err)
			}

			// Generate salt
			salt, err := generateSalt()
			if err != nil {
				return fmt.Errorf("failed to generate salt: %w", err)
			}

			masterKey := app.Encryption.DeriveKey(masterPass, salt)

			if err := app.Config.SetMasterKey(masterKey, salt); err != nil {
				return fmt.Errorf("failed to set master key: %w", err)
			}

			if err := app.Storage.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			fmt.Println("Passio initialized successfully!!")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinitialization")
	return cmd
}

func getMasterPassword() (string, error) {
	fmt.Print("Enter master password: ")
	masterPass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()

	fmt.Print("Confirm master password: ")
	confirmPass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()

	if string(masterPass) != string(confirmPass) {
		return "", fmt.Errorf("passwords do not match")
	}

	if len(masterPass) < 8 {
		return "", fmt.Errorf("master password must be at least 8 characters long")
	}

	return string(masterPass), nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}
