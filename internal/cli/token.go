package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/mahimsafa/kudo/internal/auth"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage cluster join tokens",
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate a new join token",
	RunE:  runTokenCreate,
}

var tokenTTL string

func init() {
	tokenCreateCmd.Flags().StringVar(&tokenTTL, "ttl", "24h", "Token time-to-live")
	tokenCmd.AddCommand(tokenCreateCmd)
	rootCmd.AddCommand(tokenCmd)
}

func runTokenCreate(cmd *cobra.Command, args []string) error {
	ttl, err := time.ParseDuration(tokenTTL)
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}

	secret := []byte("cluster-secret-placeholder")

	token, err := auth.GenerateJoinToken(secret, ttl)
	if err != nil {
		return err
	}

	fmt.Println("Join token (expires in", tokenTTL, "):")
	fmt.Println(token)
	fmt.Println("\nTo join a node to this cluster, run:")
	fmt.Printf("  kudo agent --join <this-node-ip>:7946 --token %s\n", token)
	return nil
}
