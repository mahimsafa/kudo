package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove applications defined in a config file",
	Long:  "Stop workloads and delete applications from cluster state. Fails without changes if another app still depends on shared routing.",
	RunE:  runRemove,
}

var removeFile string

func init() {
	removeCmd.Flags().StringVarP(&removeFile, "file", "f", "", "Path to YAML config file (required)")
	removeCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(removeFile)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	conn, err := grpcConnect()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKudoAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.Remove(ctx, &pb.RemoveRequest{YamlContent: string(data)})
	if err != nil {
		return fmt.Errorf("remove failed: %w", wrapGRPCError(err))
	}

	if !resp.Success {
		fmt.Println("Remove failed:", resp.Message)
		for _, b := range resp.Blockers {
			fmt.Printf("  - %s blocked by %s: %s (%s)\n", b.AppName, b.DependentApp, b.Reason, b.SharedResource)
		}
		return fmt.Errorf("remove blocked")
	}

	fmt.Println("Removed successfully:", resp.Message)
	return nil
}
