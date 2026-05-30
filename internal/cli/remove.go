package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove applications defined in a config file",
	Long:  "Stop workloads and delete applications from cluster state. Fails without changes if another app still depends on shared routing. Prompts for confirmation when containers are already gone.",
	RunE:  runRemove,
}

var (
	removeFile string
	removeYes  bool
)

func init() {
	removeCmd.Flags().StringVarP(&removeFile, "file", "f", "", "Path to YAML config file (required)")
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "Skip confirmation when workloads are missing")
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

	return removeWithConfirmation(ctx, client, string(data), removeYes)
}

func removeWithConfirmation(ctx context.Context, client pb.KudoAPIClient, yamlContent string, autoYes bool) error {
	resp, err := client.Remove(ctx, &pb.RemoveRequest{
		YamlContent:           yamlContent,
		ForceMissingWorkloads: false,
	})
	if err != nil {
		return fmt.Errorf("remove failed: %w", wrapGRPCError(err))
	}

	if resp.GetConfirmRequired() {
		printRemoveWarnings(resp.GetWarnings())
		if !autoYes && !promptRemoveContinue() {
			fmt.Println("Remove cancelled; no resources were changed.")
			return nil
		}

		resp, err = client.Remove(ctx, &pb.RemoveRequest{
			YamlContent:           yamlContent,
			ForceMissingWorkloads: true,
		})
		if err != nil {
			return fmt.Errorf("remove failed: %w", wrapGRPCError(err))
		}
	}

	if !resp.GetSuccess() {
		fmt.Println("Remove failed:", resp.GetMessage())
		for _, b := range resp.GetBlockers() {
			fmt.Printf("  - %s blocked by %s: %s (%s)\n", b.AppName, b.DependentApp, b.Reason, b.SharedResource)
		}
		return fmt.Errorf("remove blocked")
	}

	for _, w := range resp.GetWarnings() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}
	fmt.Println("Removed successfully:", resp.GetMessage())
	return nil
}

func printRemoveWarnings(warnings []string) {
	fmt.Fprintln(os.Stderr, "warning: some Docker containers or workloads are already missing:")
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "  - %s\n", w)
	}
}

func promptRemoveContinue() bool {
	fmt.Fprint(os.Stderr, "Remove cluster state, routes, and any remaining resources anyway? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}
