package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mahimsafa/kudo/internal/agent"
	"github.com/mahimsafa/kudo/internal/config"
	kudolog "github.com/mahimsafa/kudo/internal/log"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the Kudo agent",
	Long:  "Start the Kudo agent which participates in the cluster, manages workloads, and serves the API.",
	RunE:  runAgent,
}

var (
	agentConfigFile string
	agentBootstrap  bool
	agentJoinAddrs  []string
	agentJoinToken  string
	agentNodeName   string
)

func init() {
	agentCmd.Flags().StringVarP(&agentConfigFile, "config", "c", "/etc/kudo/kudo.yaml", "Path to config file")
	agentCmd.Flags().BoolVar(&agentBootstrap, "bootstrap", false, "Bootstrap a new cluster")
	agentCmd.Flags().StringSliceVar(&agentJoinAddrs, "join", nil, "Addresses of existing cluster nodes to join")
	agentCmd.Flags().StringVar(&agentJoinToken, "token", "", "Join token for cluster authentication")
	agentCmd.Flags().StringVar(&agentNodeName, "name", "", "Node name (defaults to hostname)")

	rootCmd.AddCommand(agentCmd)
}

func runAgent(cmd *cobra.Command, args []string) error {
	cfg, err := loadOrDefaultConfig()
	if err != nil {
		return err
	}

	if agentBootstrap {
		cfg.Cluster.Bootstrap = true
	}
	if len(agentJoinAddrs) > 0 {
		cfg.Cluster.JoinAddrs = agentJoinAddrs
	}
	if agentJoinToken != "" {
		cfg.Cluster.JoinToken = agentJoinToken
	}
	if agentNodeName != "" {
		cfg.Node.Name = agentNodeName
	}

	if cfg.Node.Name == "" {
		hostname, _ := os.Hostname()
		cfg.Node.Name = hostname
	}

	logger, err := kudolog.NewLogger(cfg.Log.Level)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer logger.Sync()

	a := agent.New(cfg, logger)
	if err := a.Start(context.Background()); err != nil {
		return fmt.Errorf("starting agent: %w", err)
	}

	a.WaitForShutdown()
	return nil
}

func loadOrDefaultConfig() (*config.AgentConfig, error) {
	if _, err := os.Stat(agentConfigFile); err == nil {
		return config.LoadAgentConfig(agentConfigFile)
	}
	return &config.AgentConfig{
		Node:    config.NodeConfig{BindAddr: "0.0.0.0", BindPort: 7946, DataDir: "/var/lib/kudo"},
		Cluster: config.ClusterConfig{},
		API:     config.APIConfig{GRPCPort: 9090, HTTPPort: 8080},
		Proxy:   config.ProxyConfig{HTTPPort: 80, HTTPSPort: 443},
		Log:     config.LogConfig{Level: "info"},
	}, nil
}
