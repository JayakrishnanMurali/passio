package cmd

import (
	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func NewRootCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "passio",
		Short: "Passio is a simple, secure, and modern password manager.",
		Long:  "Passio helps you manage your passwords easily with strong encryption. Store your passwords securely and access them from anywhere.",
	}

	cmd.AddCommand(
		newInitCmd(app),
	)

	return cmd
}