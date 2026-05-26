package gossip

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewGossipLayer(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := Config{
		NodeName: "test-node-1",
		BindAddr: "127.0.0.1",
		BindPort: 0,
	}

	g, err := NewGossipLayer(cfg, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer g.Shutdown()

	if g.NumMembers() != 1 {
		t.Errorf("expected 1 member, got %d", g.NumMembers())
	}
}

func TestGossipJoin(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg1 := Config{
		NodeName: "node-1",
		BindAddr: "127.0.0.1",
		BindPort: 0,
	}
	g1, err := NewGossipLayer(cfg1, logger)
	if err != nil {
		t.Fatalf("unexpected error creating node 1: %v", err)
	}
	defer g1.Shutdown()

	cfg2 := Config{
		NodeName:  "node-2",
		BindAddr:  "127.0.0.1",
		BindPort:  0,
		JoinAddrs: []string{g1.LocalAddr()},
	}
	g2, err := NewGossipLayer(cfg2, logger)
	if err != nil {
		t.Fatalf("unexpected error creating node 2: %v", err)
	}
	defer g2.Shutdown()

	time.Sleep(100 * time.Millisecond)

	if g1.NumMembers() != 2 {
		t.Errorf("node 1: expected 2 members, got %d", g1.NumMembers())
	}
	if g2.NumMembers() != 2 {
		t.Errorf("node 2: expected 2 members, got %d", g2.NumMembers())
	}
}
