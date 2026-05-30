package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration file",
	Long:  "Create or update applications from a YAML configuration file.",
	RunE:  runApply,
}

var applyFile string

func init() {
	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "Path to YAML config file (required)")
	applyCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(applyCmd)
}

func runApply(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(applyFile)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	conn, err := grpcConnect()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKudoAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Apply(ctx, &pb.ApplyRequest{YamlContent: string(data)})
	if err != nil {
		return fmt.Errorf("apply failed: %w", wrapGRPCError(err))
	}

	if resp.Success {
		fmt.Println("Applied successfully:", resp.Message)
		return nil
	}
	return fmt.Errorf("apply failed: %s", resp.Message)
}
