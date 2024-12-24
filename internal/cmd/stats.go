package cmd

import (
	"fmt"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

func newStatsCmd(app *app.App) *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show password manager statistics",
		Long: `Display statistics about stored passwords including:
- Total number of entries
- Password age information
- Security statistics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			stats, err := app.Storage.GetStats()
			if err != nil {
				return fmt.Errorf("failed to get statistics: %w", err)
			}

			// Print basic stats
			fmt.Println("Password Manager Statistics")
			fmt.Println("-------------------------")
			fmt.Printf("Total entries: %d\n", stats.TotalEntries)

			if stats.TotalEntries > 0 {
				fmt.Printf("Oldest entry: %s\n", stats.OldestEntry.Format("2006-01-02"))
				fmt.Printf("Newest entry: %s\n", stats.NewestEntry.Format("2006-01-02"))
				fmt.Printf("Average password age: %.1f days\n", stats.AveragePassAge)

				if detailed {
					// Get and analyze all entries for detailed stats
					entries, err := app.Storage.ListEntries()
					if err != nil {
						return fmt.Errorf("failed to list entries: %w", err)
					}

					var (
						expiredCount    int
						weakCount       int
						reusedPasswords = make(map[string][]string)
					)

					for _, entry := range entries {
						// Check expired passwords
						age := time.Since(entry.UpdatedAt).Hours() / 24
						if age > float64(app.Config.PasswordExpiration) {
							expiredCount++
						}

						// Decrypt and check password strength
						password, err := app.DecryptPassword(entry.Password)
						if err != nil {
							return fmt.Errorf("failed to decrypt password: %w", err)
						}

						health := app.CheckPasswordHealth(password)
						if !health["length"] || !health["uppercase"] ||
							!health["lowercase"] || !health["numbers"] ||
							!health["specialChars"] || !health["notCommon"] {
							weakCount++
						}

						// Track password reuse
						reusedPasswords[password] = append(reusedPasswords[password], entry.Name)
					}

					fmt.Println("\nDetailed Statistics")
					fmt.Println("-------------------")
					fmt.Printf("Expired passwords: %d\n", expiredCount)
					fmt.Printf("Weak passwords: %d\n", weakCount)

					// Report password reuse
					var reusedCount int
					for _, entries := range reusedPasswords {
						if len(entries) > 1 {
							reusedCount++
						}
					}
					fmt.Printf("Reused passwords: %d\n", reusedCount)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed statistics")

	return cmd
}
