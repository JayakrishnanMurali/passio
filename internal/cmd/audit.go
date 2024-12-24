package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newAuditCmd(app *app.App) *cobra.Command {
	var (
		checkWeak    bool
		checkReused  bool
		checkExpired bool
		verbose      bool
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit password security",
		Long: `Audit password security by checking for:
- Weak passwords (less than required length, missing character types)
- Reused passwords across different entries
- Expired passwords (older than configured expiration period)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			// Get all entries
			entries, err := app.Storage.ListEntries()
			if err != nil {
				return fmt.Errorf("failed to list entries: %w", err)
			}

			var issues []string
			passwordMap := make(map[string][]string) // For checking reused passwords

			// Check each entry
			for _, entry := range entries {
				// Decrypt password for checking
				password, err := app.DecryptPassword(entry.Password)
				if err != nil {
					return fmt.Errorf("failed to decrypt password for entry %s: %w", entry.Name, err)
				}

				// Check weak passwords
				if checkWeak {
					health := app.CheckPasswordHealth(password)
					var weaknesses []string

					if !health["length"] {
						weaknesses = append(weaknesses, "too short")
					}
					if !health["uppercase"] {
						weaknesses = append(weaknesses, "no uppercase")
					}
					if !health["lowercase"] {
						weaknesses = append(weaknesses, "no lowercase")
					}
					if !health["numbers"] {
						weaknesses = append(weaknesses, "no numbers")
					}
					if !health["specialChars"] {
						weaknesses = append(weaknesses, "no special characters")
					}
					if !health["notCommon"] {
						weaknesses = append(weaknesses, "common password")
					}

					if len(weaknesses) > 0 {
						issue := fmt.Sprintf("Weak password for %s: %s",
							entry.Name, strings.Join(weaknesses, ", "))
						issues = append(issues, issue)
					}
				}

				// Track passwords for reuse checking
				if checkReused {
					passwordMap[password] = append(passwordMap[password], entry.Name)
				}

				// Check expired passwords
				if checkExpired && app.Config.PasswordExpiration > 0 {
					age := time.Since(entry.UpdatedAt).Hours() / 24
					if age > float64(app.Config.PasswordExpiration) {
						issue := fmt.Sprintf("Expired password for %s (%.0f days old)",
							entry.Name, age)
						issues = append(issues, issue)
					}
				}
			}

			// Check for reused passwords
			if checkReused {
				for _, entries := range passwordMap {
					if len(entries) > 1 {
						sort.Strings(entries)
						issue := fmt.Sprintf("Password reused across entries: %s",
							strings.Join(entries, ", "))
						issues = append(issues, issue)
					}
				}
			}

			// Print results
			if len(issues) == 0 {
				fmt.Println("No issues found!")
				return nil
			}

			fmt.Printf("Found %d issues:\n", len(issues))
			for i, issue := range issues {
				if verbose {
					fmt.Printf("%d. %s\n", i+1, issue)
				} else {
					// Print shortened version for non-verbose output
					parts := strings.SplitN(issue, ":", 2)
					fmt.Printf("%d. %s\n", i+1, parts[0])
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&checkWeak, "weak", "w", true, "Check for weak passwords")
	cmd.Flags().BoolVarP(&checkReused, "reused", "r", true, "Check for reused passwords")
	cmd.Flags().BoolVarP(&checkExpired, "expired", "e", true, "Check for expired passwords")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed issue descriptions")

	return cmd
}
