package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/jayakrishnanMurali/passio/internal/storage"
	"github.com/spf13/cobra"
)

func newSearchCmd(app *app.App) *cobra.Command {
	var (
		showTags bool
		byTag    bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for password entries",
		Long: `Search for password entries by name, username, URL, or tags.
Use --by-tag to search only in tags.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			query := args[0]
			var entries []*storage.Entry
			var err error

			if byTag {
				entries, err = app.Storage.GetEntriesByTag(query)
			} else {
				entries, err = app.Storage.SearchEntries(query)
			}

			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("No matching entries found")
				return nil
			}

			// Create tabwriter for formatted output
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Print header
			headers := []string{"Name", "Username", "URL", "Last Modified"}
			if showTags {
				headers = append(headers, "Tags")
			}
			fmt.Fprintln(w, strings.Join(headers, "\t"))
			fmt.Fprintln(w, strings.Repeat("-", 80))

			// Print entries
			for _, entry := range entries {
				row := []string{
					entry.Name,
					entry.Username,
					entry.URL,
					entry.UpdatedAt.Format("2006-01-02 15:04:05"),
				}

				if showTags {
					row = append(row, strings.Join(entry.Tags, ", "))
				}

				fmt.Fprintln(w, strings.Join(row, "\t"))
			}

			w.Flush()
			fmt.Printf("\nFound %d matching entries\n", len(entries))
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&showTags, "show-tags", "t", false, "Show tags in results")
	cmd.Flags().BoolVarP(&byTag, "by-tag", "b", false, "Search only in tags")

	return cmd
}
