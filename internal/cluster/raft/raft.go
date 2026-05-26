package raftlayer

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/mahimsafa/kudo/internal/cluster/state"
)

type Config struct {
	NodeID    string
	BindAddr  string
	DataDir   string
	Bootstrap bool
}

type RaftNode struct {
	raft *raft.Raft
	fsm  *state.FSM
}

func NewRaftNode(cfg Config) (*RaftNode, error) {
	fsm := state.NewFSM()

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.NodeID)

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-log.db"))
	if err != nil {
		return nil, fmt.Errorf("creating log store: %w", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-stable.db"))
	if err != nil {
		return nil, fmt.Errorf("creating stable store: %w", err)
	}

	snapshotStore := raft.NewDiscardSnapshotStore()

	addr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("resolving bind addr: %w", err)
	}

	transport, err := raft.NewTCPTransport(cfg.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("creating transport: %w", err)
	}

	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("creating raft: %w", err)
	}

	if cfg.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(cfg.NodeID),
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(configuration)
	}

	return &RaftNode{raft: r, fsm: fsm}, nil
}

func (n *RaftNode) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

func (n *RaftNode) LeaderAddr() string {
	addr, _ := n.raft.LeaderWithID()
	return string(addr)
}

func (n *RaftNode) Apply(data []byte, timeout time.Duration) error {
	future := n.raft.Apply(data, timeout)
	if err := future.Error(); err != nil {
		return err
	}
	if resp := future.Response(); resp != nil {
		if err, ok := resp.(error); ok {
			return err
		}
	}
	return nil
}

func (n *RaftNode) AddVoter(id, addr string) error {
	future := n.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 10*time.Second)
	return future.Error()
}

func (n *RaftNode) RemoveServer(id string) error {
	future := n.raft.RemoveServer(raft.ServerID(id), 0, 10*time.Second)
	return future.Error()
}

func (n *RaftNode) FSM() *state.FSM {
	return n.fsm
}

func (n *RaftNode) Shutdown() error {
	return n.raft.Shutdown().Error()
}
