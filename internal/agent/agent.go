package agent

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mahimsafa/kudo/internal/api"
	"github.com/mahimsafa/kudo/internal/cluster/gossip"
	raftlayer "github.com/mahimsafa/kudo/internal/cluster/raft"
	"github.com/mahimsafa/kudo/internal/config"
	"github.com/mahimsafa/kudo/internal/executor"
	"github.com/mahimsafa/kudo/internal/executor/docker"
	"github.com/mahimsafa/kudo/internal/proxy"
	"github.com/mahimsafa/kudo/internal/reconciler"
)

type Agent struct {
	config    *config.AgentConfig
	logger    *zap.Logger
	gossip    *gossip.GossipLayer
	raft      *raftlayer.RaftNode
	api       *api.Server
	executor  *executor.Executor
	proxy     *proxy.Proxy
	cancel    context.CancelFunc
}

func New(cfg *config.AgentConfig, logger *zap.Logger) *Agent {
	return &Agent{
		config: cfg,
		logger: logger,
	}
}

func (a *Agent) Start(ctx context.Context) error {
	a.logger.Info("starting kudo agent", zap.String("node", a.config.Node.Name))

	gossipCfg := gossip.Config{
		NodeName:  a.config.Node.Name,
		BindAddr:  a.config.Node.BindAddr,
		BindPort:  a.config.Node.BindPort,
		JoinAddrs: a.config.Cluster.JoinAddrs,
	}
	g, err := gossip.NewGossipLayer(gossipCfg, a.logger)
	if err != nil {
		return fmt.Errorf("starting gossip: %w", err)
	}
	a.gossip = g

	raftCfg := raftlayer.Config{
		NodeID:    a.config.Node.Name,
		BindAddr:  fmt.Sprintf("%s:%d", a.config.Node.BindAddr, a.config.API.GRPCPort+1000),
		DataDir:   a.config.Node.DataDir,
		Bootstrap: a.config.Cluster.Bootstrap,
	}
	r, err := raftlayer.NewRaftNode(raftCfg)
	if err != nil {
		g.Shutdown()
		return fmt.Errorf("starting raft: %w", err)
	}
	a.raft = r

	a.executor = executor.NewExecutor(a.logger)
	dockerAdapter, err := docker.NewDockerAdapter(a.logger)
	if err != nil {
		a.logger.Warn("docker adapter unavailable, continuing without it", zap.Error(err))
	} else {
		a.executor.RegisterAdapter(dockerAdapter)
	}

	grpcAddr := net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", a.config.API.GRPCPort))
	a.api = api.NewServer(a.raft, a.logger)
	if err := a.api.Start(grpcAddr); err != nil {
		a.Shutdown()
		return fmt.Errorf("starting API server: %w", err)
	}

	a.proxy = proxy.NewProxy()
	proxyAddr := fmt.Sprintf(":%d", a.config.Proxy.HTTPPort)
	go func() {
		a.logger.Info("L7 proxy starting", zap.String("addr", proxyAddr))
		if err := a.proxy.ListenAndServe(proxyAddr); err != nil {
			a.logger.Error("proxy server error", zap.Error(err))
		}
	}()

	runCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	reconcileLoop := reconciler.NewReconcileLoop(a.raft, a.logger, 10*time.Second, a.handleReconcileAction)
	go reconcileLoop.Start(runCtx)

	a.logger.Info("kudo agent started successfully",
		zap.Int("gossip_members", a.gossip.NumMembers()),
		zap.Bool("raft_leader", a.raft.IsLeader()),
		zap.Int("grpc_port", a.config.API.GRPCPort),
		zap.Int("proxy_port", a.config.Proxy.HTTPPort),
	)

	return nil
}

func (a *Agent) handleReconcileAction(action reconciler.Action) error {
	switch action.Type {
	case reconciler.ActionDeploy:
		a.logger.Info("deploy action",
			zap.String("app", action.AppName),
			zap.String("node", action.NodeID),
		)
	case reconciler.ActionStop:
		if a.executor != nil {
			return a.executor.Stop(context.Background(), action.Adapter, executor.StopRequest{
				InstanceID: action.InstanceID,
			})
		}
	}
	return nil
}

func (a *Agent) WaitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	a.logger.Info("shutdown signal received")
	a.Shutdown()
}

func (a *Agent) Shutdown() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.api != nil {
		a.api.Stop()
	}
	if a.raft != nil {
		a.raft.Shutdown()
	}
	if a.gossip != nil {
		a.gossip.Shutdown()
	}
	a.logger.Info("agent shut down")
}

func (a *Agent) Raft() *raftlayer.RaftNode {
	return a.raft
}

func (a *Agent) Gossip() *gossip.GossipLayer {
	return a.gossip
}
