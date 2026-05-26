package gossip

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

type Config struct {
	NodeName  string
	BindAddr  string
	BindPort  int
	JoinAddrs []string
}

type GossipLayer struct {
	list   *memberlist.Memberlist
	logger *zap.Logger
	events *eventDelegate
}

func NewGossipLayer(cfg Config, logger *zap.Logger) (*GossipLayer, error) {
	events := &eventDelegate{logger: logger}

	mlCfg := memberlist.DefaultLANConfig()
	mlCfg.Name = cfg.NodeName
	mlCfg.BindAddr = cfg.BindAddr
	mlCfg.BindPort = cfg.BindPort
	mlCfg.Events = events
	mlCfg.LogOutput = nil

	list, err := memberlist.Create(mlCfg)
	if err != nil {
		return nil, fmt.Errorf("creating memberlist: %w", err)
	}

	g := &GossipLayer{
		list:   list,
		logger: logger,
		events: events,
	}

	if len(cfg.JoinAddrs) > 0 {
		n, err := list.Join(cfg.JoinAddrs)
		if err != nil {
			list.Shutdown()
			return nil, fmt.Errorf("joining cluster: %w", err)
		}
		logger.Info("joined cluster", zap.Int("nodes_contacted", n))
	}

	return g, nil
}

func (g *GossipLayer) NumMembers() int {
	return g.list.NumMembers()
}

func (g *GossipLayer) Members() []*memberlist.Node {
	return g.list.Members()
}

func (g *GossipLayer) LocalAddr() string {
	node := g.list.LocalNode()
	return net.JoinHostPort(node.Addr.String(), strconv.Itoa(int(node.Port)))
}

func (g *GossipLayer) OnJoin(fn func(node *memberlist.Node)) {
	g.events.onJoin = fn
}

func (g *GossipLayer) OnLeave(fn func(node *memberlist.Node)) {
	g.events.onLeave = fn
}

func (g *GossipLayer) Shutdown() error {
	return g.list.Shutdown()
}
