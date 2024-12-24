package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var (
		length      int
		special     bool
		numbers     bool
		uppercase   bool
		lowercase   bool
		noAmbiguous bool
		copy        bool
		count       int
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a random password",
		Long: `Generate one or more random passwords with specified options.
By default, generates a single password with all character types enabled.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if length < 1 {
				return fmt.Errorf("password length must be positive")
			}

			if !special && !numbers && !uppercase && !lowercase {
				// If no character types specified, enable all
				special = true
				numbers = true
				uppercase = true
				lowercase = true
			}

			for i := 0; i < count; i++ {
				password, err := generatePasswordWithOptions(length, special, numbers, uppercase, lowercase, noAmbiguous)
				if err != nil {
					return fmt.Errorf("failed to generate password: %w", err)
				}

				if copy && i == 0 {
					if err := clipboard.WriteAll(password); err != nil {
						return fmt.Errorf("failed to copy to clipboard: %w", err)
					}
					fmt.Println("Password copied to clipboard")
				}

				fmt.Println(password)
			}

			return nil

		},
	}

	cmd.Flags().IntVarP(&length, "length", "l", 16, "Length of generated password")
	cmd.Flags().BoolVarP(&special, "special", "s", true, "Include special characters")
	cmd.Flags().BoolVarP(&numbers, "numbers", "n", true, "Include numbers")
	cmd.Flags().BoolVarP(&uppercase, "uppercase", "u", true, "Include uppercase letters")
	cmd.Flags().BoolVarP(&lowercase, "lowercase", "w", true, "Include lowercase letters")
	cmd.Flags().BoolVar(&noAmbiguous, "no-ambiguous", false, "Exclude ambiguous characters (1/l, 0/O, etc.)")
	cmd.Flags().BoolVarP(&copy, "copy", "c", false, "Copy first generated password to clipboard")
	cmd.Flags().IntVarP(&count, "count", "t", 1, "Number of passwords to generate")

	return cmd
}

func generatePassword(length int, special bool) (string, error) {
	return generatePasswordWithOptions(length, special, true, true, true, false)
}

func generatePasswordWithOptions(length int, special, numbers, uppercase, lowercase, noAmbiguous bool) (string, error) {
	var chars string

	if uppercase {
		if noAmbiguous {
			chars += "ABCDEFGHJKLMNPQRSTUVWXYZ"
		} else {
			chars += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		}
	}

	if lowercase {
		if noAmbiguous {
			chars += "abcdefghijkmnpqrstuvwxyz"
		} else {
			chars += "abcdefghijklmnopqrstuvwxyz"
		}
	}

	if numbers {
		if noAmbiguous {
			chars += "23456789"
		} else {
			chars += "0123456789"
		}
	}

	if special {
		chars += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	if chars == "" {
		return "", fmt.Errorf("no character sets selected")
	}

	var password strings.Builder
	password.Grow(length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password.WriteByte(chars[n.Int64()])
	}

	result := password.String()
	if !validatePassword(result, special, numbers, uppercase, lowercase) {
		// If validation fails, generate a new password
		return generatePasswordWithOptions(length, special, numbers, uppercase, lowercase, noAmbiguous)
	}

	return result, nil
}

func validatePassword(password string, special, numbers, uppercase, lowercase bool) bool {
	hasSpecial := !special
	hasNumber := !numbers
	hasUpper := !uppercase
	hasLower := !lowercase

	for _, c := range password {
		if special && strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", c) {
			hasSpecial = true
		}
		if numbers && strings.ContainsRune("0123456789", c) {
			hasNumber = true
		}
		if uppercase && c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if lowercase && c >= 'a' && c <= 'z' {
			hasLower = true
		}
	}

	return hasSpecial && hasNumber && hasUpper && hasLower
}
