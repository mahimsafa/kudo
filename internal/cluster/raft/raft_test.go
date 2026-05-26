package raftlayer

import (
	"testing"
	"time"
)

func TestNewRaftNode(t *testing.T) {
	cfg := Config{
		NodeID:    "test-node-1",
		BindAddr:  "127.0.0.1:0",
		DataDir:   t.TempDir(),
		Bootstrap: true,
	}

	node, err := NewRaftNode(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer node.Shutdown()

	time.Sleep(3 * time.Second)

	if !node.IsLeader() {
		t.Error("expected single bootstrap node to be leader")
	}
}
