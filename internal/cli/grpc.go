package cli

import (
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const defaultGRPCServer = "127.0.0.1:9090"

var grpcServerFlag string

func init() {
	rootCmd.PersistentFlags().StringVar(&grpcServerFlag, "server", "",
		"gRPC API address (default 127.0.0.1:9090, or KUDO_SERVER env)")
}

func grpcServerAddress() string {
	if grpcServerFlag != "" {
		return grpcServerFlag
	}
	if addr := os.Getenv("KUDO_SERVER"); addr != "" {
		return addr
	}
	return defaultGRPCServer
}

func grpcConnect() (*grpc.ClientConn, error) {
	addr := grpcServerAddress()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", addr, err)
	}
	return conn, nil
}

func wrapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.Unavailable {
		return fmt.Errorf("%w: no agent listening on %s — start one in another terminal (e.g. kudo agent --bootstrap --name dev-node-1)",
			err, grpcServerAddress())
	}
	return err
}
