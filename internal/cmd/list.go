package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/jayakrishnanMurali/passio/internal/storage"
	"github.com/spf13/cobra"
)

func newListCmd(app *app.App) *cobra.Command {
	var (
		filter   string
		sortBy   string
		showAll  bool
		showTags bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all password entries",
		Long: `List all password entries in a tabular format.
Entries can be filtered and sorted based on various criteria.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("passio is locked. Please unlock first")
			}

			entries, err := app.Storage.ListEntries()
			if err != nil {
				return fmt.Errorf("failed to list entries: %w", err)
			}

			if filter != "" {
				filtered := make([]*storage.Entry, 0)
				filterLower := strings.ToLower(filter)

				for _, entry := range entries {
					if strings.Contains(strings.ToLower(entry.Name), filterLower) ||
						strings.Contains(strings.ToLower(entry.Username), filterLower) ||
						strings.Contains(strings.ToLower(entry.URL), filterLower) ||
						(showTags && containsTag(entry.Tags, filterLower)) {
						filtered = append(filtered, entry)
					}
				}

				entries = filtered
			}

			sortEntries(entries, sortBy)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			headers := []string{"Name", "Username", "URL", "Created", "Last Modified"}
			if showTags {
				headers = append(headers, "Tags")
			}
			fmt.Fprintln(w, strings.Join(headers, "\t"))
			fmt.Fprintln(w, strings.Repeat("-", 80))

			for _, entry := range entries {

				// Format dates
				created := entry.CreatedAt.Format("2006-01-02")
				modified := entry.UpdatedAt.Format("2006-01-02")

				// Check password age
				passwordAge := time.Since(entry.UpdatedAt).Hours() / 24
				ageIndicator := " "
				if passwordAge > float64(app.Config.PasswordExpiration) {
					ageIndicator = "!" // Indicate old password
				}

				// Format row
				row := []string{
					ageIndicator + entry.Name,
					entry.Username,
					entry.URL,
					created,
					modified,
				}

				if showTags {
					row = append(row, strings.Join(entry.Tags, ", "))
				}

				fmt.Fprintln(w, strings.Join(row, "\t"))
			}

			w.Flush()

			fmt.Printf("\nTotal entries: %d\n", len(entries))
			if app.Config.PasswordExpiration > 0 {
				fmt.Println("! indicates password older than configured expiration period")
			}

			return nil

		},
	}

	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter entries by name, username, or URL")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Sort entries by: name, username, created, modified")
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all entry details")
	cmd.Flags().BoolVarP(&showTags, "tags", "t", false, "Show entry tags")

	return cmd
}

func containsTag(tags []string, search string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), search) {
			return true
		}
	}
	return false
}

func sortEntries(entries []*storage.Entry, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "username":
		sortByUsername(entries)
	case "created":
		sortByCreated(entries)
	case "modified":
		sortByModified(entries)
	default:
		sortByName(entries)
	}
}

func sortByName(entries []*storage.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

func sortByUsername(entries []*storage.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Username) < strings.ToLower(entries[j].Username)
	})
}

func sortByCreated(entries []*storage.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})
}

func sortByModified(entries []*storage.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.Before(entries[j].UpdatedAt)
	})
}
