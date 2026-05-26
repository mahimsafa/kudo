package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
)

var scaleCmd = &cobra.Command{
	Use:   "scale <app-name>",
	Short: "Scale an application",
	Args:  cobra.ExactArgs(1),
	RunE:  runScale,
}

var scaleReplicas int32

func init() {
	scaleCmd.Flags().Int32Var(&scaleReplicas, "replicas", 0, "Number of replicas (required)")
	scaleCmd.MarkFlagRequired("replicas")
	rootCmd.AddCommand(scaleCmd)
}

func runScale(cmd *cobra.Command, args []string) error {
	conn, err := grpcConnect()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKudoAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ScaleApplication(ctx, &pb.ScaleRequest{
		AppName:  args[0],
		Replicas: scaleReplicas,
	})
	if err != nil {
		return err
	}

	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		fmt.Println("Scale failed:", resp.Message)
	}
	return nil
}
