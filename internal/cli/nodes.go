package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List cluster nodes",
	RunE:  runNodes,
}

func init() {
	rootCmd.AddCommand(nodesCmd)
}

func runNodes(cmd *cobra.Command, args []string) error {
	conn, err := grpcConnect()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKudoAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListNodes(ctx, &pb.ListNodesRequest{})
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tADDRESS\tSTATUS\n")
	for _, node := range resp.Nodes {
		id := node.Id
		if len(id) > 8 {
			id = id[:8]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, node.Name, node.Address, node.Status)
	}
	w.Flush()
	return nil
}
