package genkey

import (
	"fmt"
	"os"

	"github.com/lore/goober/internal/crypto"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "genkey",
		Short: "Generate ed25519 key pair",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate key pair
			keyPair, err := crypto.GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair: %w", err)
			}

			// Set output directory
			if outputDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to find home directory: %w", err)
				}
				outputDir = fmt.Sprintf("%s/.config/goober", home)
			}

			// Ensure directory exists
			if err := os.MkdirAll(outputDir, 0700); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Write private key
			privateKeyPath := fmt.Sprintf("%s/key", outputDir)
			if err := keyPair.PrivateKey.WriteFile(privateKeyPath); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}

			// Write public key
			publicKeyPath := fmt.Sprintf("%s/key.pub", outputDir)
			if err := keyPair.PublicKey.WriteFile(publicKeyPath); err != nil {
				return fmt.Errorf("failed to write public key: %w", err)
			}

			// Set proper permissions
			if err := os.Chmod(privateKeyPath, 0600); err != nil {
				return fmt.Errorf("failed to set permissions on private key: %w", err)
			}

			// Output the public key in SSH format
			fmt.Printf("Generated ed25519 key pair:\n")
			fmt.Printf("Private key: %s (0600)\n", privateKeyPath)
			fmt.Printf("Public key: %s\n", publicKeyPath)
			fmt.Printf("\nPublic key (for config): %s\n", keyPair.PublicKey.String())

			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to write key files to (default: ~/.config/goober)")

	return cmd
}