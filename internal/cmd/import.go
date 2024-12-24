package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/jayakrishnanMurali/passio/internal/storage"
	"github.com/spf13/cobra"
)

func newImportCmd(app *app.App) *cobra.Command {
	var (
		format   string
		decrypt  bool
		dryRun   bool
		skipDups bool
	)

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import password entries",
		Long: `Import password entries from a JSON or CSV file.
Supports importing encrypted or decrypted passwords.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.IsLocked() {
				return fmt.Errorf("password manager is locked. Please unlock first")
			}

			filename := args[0]
			if _, err := os.Stat(filename); err != nil {
				return fmt.Errorf("import file not found: %w", err)
			}

			var importedData *ExportData
			var err error

			// Import based on format
			switch format {
			case "json":
				importedData, err = importJSON(filename)
			case "csv":
				importedData, err = importCSV(filename)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("failed to import data: %w", err)
			}

			// Process entries
			var imported, skipped int
			for _, importEntry := range importedData.Entries {
				// Check if entry already exists
				existing, err := app.Storage.GetEntry(importEntry.Name)
				if err == nil && existing != nil {
					if skipDups {
						skipped++
						continue
					}
					return fmt.Errorf("entry already exists: %s", importEntry.Name)
				}

				// Create new entry
				entry := &storage.Entry{
					Name:      importEntry.Name,
					Username:  importEntry.Username,
					URL:       importEntry.URL,
					Notes:     importEntry.Notes,
					Tags:      importEntry.Tags,
					CreatedAt: importEntry.CreatedAt,
					UpdatedAt: importEntry.UpdatedAt,
				}

				// Handle password
				if importedData.Encrypted {
					entry.Password = importEntry.Password
				} else {
					// Encrypt password if it was imported in plain text
					encryptedPass, err := app.EncryptPassword(string(importEntry.Password))
					if err != nil {
						return fmt.Errorf("failed to encrypt password for entry %s: %w", entry.Name, err)
					}
					entry.Password = encryptedPass
				}

				// Add entry unless this is a dry run
				if !dryRun {
					if err := app.Storage.AddEntry(entry); err != nil {
						return fmt.Errorf("failed to add entry %s: %w", entry.Name, err)
					}
				}
				imported++
			}

			fmt.Printf("Import summary:\n")
			fmt.Printf("- Imported: %d entries\n", imported)
			if skipped > 0 {
				fmt.Printf("- Skipped: %d duplicate entries\n", skipped)
			}
			if dryRun {
				fmt.Println("This was a dry run - no entries were actually imported")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Import format (json or csv)")
	cmd.Flags().BoolVarP(&decrypt, "decrypt", "d", false, "Import decrypted passwords")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate import without making changes")
	cmd.Flags().BoolVar(&skipDups, "skip-duplicates", false, "Skip duplicate entries instead of failing")

	return cmd
}

func importJSON(filename string) (*ExportData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	var data ExportData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &data, nil
}

func importCSV(filename string) (*ExportData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	data := &ExportData{
		Version:    "1.0",
		ExportDate: time.Now(),
		Encrypted:  false,
		Entries:    make([]*ExportEntry, 0),
	}

	// Read CSV file line by line
	scanner := bufio.NewScanner(file)

	// Skip header
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Process entries
	for scanner.Scan() {
		line := scanner.Text()
		fields := parseCSVLine(line)
		if len(fields) < 8 {
			continue // Skip invalid lines
		}

		createdAt, _ := time.Parse(time.RFC3339, fields[6])
		updatedAt, _ := time.Parse(time.RFC3339, fields[7])

		entry := &ExportEntry{
			Name:      fields[0],
			Username:  fields[1],
			Password:  []byte(fields[2]),
			URL:       fields[3],
			Notes:     fields[4],
			Tags:      strings.Split(fields[5], ";"),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		data.Entries = append(data.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	return data, nil
}

// parseCSVLine parses a CSV line handling quoted fields
func parseCSVLine(line string) []string {
	var fields []string
	var field strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		char := line[i]
		switch {
		case char == '"':
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				field.WriteByte('"')
				i++
			} else {
				inQuotes = !inQuotes
			}
		case char == ',' && !inQuotes:
			fields = append(fields, field.String())
			field.Reset()
		default:
			field.WriteByte(char)
		}
	}
	fields = append(fields, field.String())
	return fields
}
