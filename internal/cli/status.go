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

var statusCmd = &cobra.Command{
	Use:   "status [app-name]",
	Short: "Show application status",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	conn, err := grpcConnect()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKudoAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if len(args) > 0 {
		resp, err := client.GetStatus(ctx, &pb.StatusRequest{AppName: args[0]})
		if err != nil {
			return err
		}
		fmt.Printf("Application: %s\n", resp.AppName)
		fmt.Printf("Adapter:     %s\n", resp.Adapter)
		fmt.Printf("Replicas:    %d/%d running\n", resp.RunningReplicas, resp.DesiredReplicas)
		fmt.Println("\nInstances:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tNODE\tSTATUS\tADDRESS\n")
		for _, inst := range resp.Instances {
			id := inst.Id
			if len(id) > 8 {
				id = id[:8]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, inst.NodeId, inst.Status, inst.Address)
		}
		w.Flush()
	} else {
		resp, err := client.ListApplications(ctx, &pb.ListAppsRequest{})
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "NAME\tADAPTER\tREPLICAS\n")
		for _, app := range resp.Apps {
			fmt.Fprintf(w, "%s\t%s\t%d\n", app.Name, app.Adapter, app.Replicas)
		}
		w.Flush()
	}
	return nil
}
