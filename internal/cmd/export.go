package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/spf13/cobra"
)

type ExportData struct {
	Version    string         `json:"version"`
	ExportDate time.Time      `json:"export_date"`
	Entries    []*ExportEntry `json:"entries"`
	Encrypted  bool           `json:"encrypted"`
}

type ExportEntry struct {
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  []byte    `json:"password"`
	URL       string    `json:"url"`
	Notes     string    `json:"notes"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newExportCmd(app *app.App) *cobra.Command {
	var (
		outputFile string
		decrypt    bool
		format     string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export password entries",
		Long: `Export password entries to a file in JSON or CSV format.
Passwords can be exported in encrypted or decrypted form.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			// Get all entries
			entries, err := app.Storage.ListEntries()
			if err != nil {
				return fmt.Errorf("failed to list entries: %w", err)
			}

			// Prepare export data
			exportData := &ExportData{
				Version:    "1.0",
				ExportDate: time.Now(),
				Encrypted:  !decrypt,
				Entries:    make([]*ExportEntry, 0, len(entries)),
			}

			// Process entries
			for _, entry := range entries {
				exportEntry := &ExportEntry{
					Name:      entry.Name,
					Username:  entry.Username,
					URL:       entry.URL,
					Notes:     entry.Notes,
					Tags:      entry.Tags,
					CreatedAt: entry.CreatedAt,
					UpdatedAt: entry.UpdatedAt,
				}

				if decrypt {
					// Decrypt password if requested
					password, err := app.DecryptPassword(entry.Password)
					if err != nil {
						return fmt.Errorf("failed to decrypt password for entry %s: %w", entry.Name, err)
					}
					exportEntry.Password = []byte(password)
				} else {
					exportEntry.Password = entry.Password
				}

				exportData.Entries = append(exportData.Entries, exportEntry)
			}

			// Create output directory if it doesn't exist
			if outputFile == "" {
				outputFile = fmt.Sprintf("pm_export_%s.%s",
					time.Now().Format("20060102_150405"), format)
			}
			if err := os.MkdirAll(filepath.Dir(outputFile), 0700); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Export based on format
			switch format {
			case "json":
				if err := exportJSON(outputFile, exportData); err != nil {
					return err
				}
			case "csv":
				if err := exportCSV(outputFile, exportData); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}

			fmt.Printf("Successfully exported %d entries to %s\n", len(entries), outputFile)
			if !decrypt {
				fmt.Println("Passwords were exported in encrypted form")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	cmd.Flags().BoolVarP(&decrypt, "decrypt", "d", false, "Export decrypted passwords (warning: sensitive!)")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Export format (json or csv)")

	return cmd
}

func exportJSON(filename string, data *ExportData) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}

func exportCSV(filename string, data *ExportData) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	// Write CSV header
	header := "Name,Username,Password,URL,Notes,Tags,Created,Updated\n"
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write entries
	for _, entry := range data.Entries {
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			escapeCSV(entry.Name),
			escapeCSV(entry.Username),
			escapeCSV(string(entry.Password)),
			escapeCSV(entry.URL),
			escapeCSV(entry.Notes),
			escapeCSV(joinTags(entry.Tags)),
			entry.CreatedAt.Format(time.RFC3339),
			entry.UpdatedAt.Format(time.RFC3339),
		)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write CSV line: %w", err)
		}
	}

	return nil
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(s, "\"", "\"\""))
	}
	return s
}

func joinTags(tags []string) string {
	return strings.Join(tags, ";")
}
