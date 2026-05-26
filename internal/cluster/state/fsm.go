package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type OpType string

const (
	OpSetApplication    OpType = "set_application"
	OpDeleteApplication OpType = "delete_application"
	OpSetNode           OpType = "set_node"
	OpDeleteNode        OpType = "delete_node"
	OpSetInstance       OpType = "set_instance"
	OpDeleteInstance    OpType = "delete_instance"
)

type Command struct {
	Op   OpType          `json:"op"`
	Data json.RawMessage `json:"data"`
}

type Application struct {
	Name     string            `json:"name"`
	Adapter  string            `json:"adapter"`
	Replicas int               `json:"replicas"`
	Spec     map[string]string `json:"spec,omitempty"`
	Routing  RoutingConfig     `json:"routing,omitempty"`
	Version  int               `json:"version"`
}

type RoutingConfig struct {
	Domain      string `json:"domain,omitempty"`
	Path        string `json:"path,omitempty"`
	TLS         string `json:"tls,omitempty"`
	Algorithm   string `json:"algorithm,omitempty"`
	HealthCheck string `json:"healthcheck,omitempty"`
}

type Node struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

type Instance struct {
	ID      string `json:"id"`
	AppName string `json:"app_name"`
	NodeID  string `json:"node_id"`
	Status  string `json:"status"`
	Address string `json:"address"`
}

type ClusterState struct {
	Applications map[string]Application `json:"applications"`
	Nodes        map[string]Node        `json:"nodes"`
	Instances    map[string]Instance    `json:"instances"`
}

type FSM struct {
	mu    sync.RWMutex
	state ClusterState
}

func NewFSM() *FSM {
	return &FSM{
		state: ClusterState{
			Applications: make(map[string]Application),
			Nodes:        make(map[string]Node),
			Instances:    make(map[string]Instance),
		},
	}
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("unmarshal command: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Op {
	case OpSetApplication:
		var app Application
		if err := json.Unmarshal(cmd.Data, &app); err != nil {
			return fmt.Errorf("unmarshal application: %w", err)
		}
		f.state.Applications[app.Name] = app

	case OpDeleteApplication:
		var name string
		if err := json.Unmarshal(cmd.Data, &name); err != nil {
			return fmt.Errorf("unmarshal app name: %w", err)
		}
		delete(f.state.Applications, name)

	case OpSetNode:
		var node Node
		if err := json.Unmarshal(cmd.Data, &node); err != nil {
			return fmt.Errorf("unmarshal node: %w", err)
		}
		f.state.Nodes[node.ID] = node

	case OpDeleteNode:
		var id string
		if err := json.Unmarshal(cmd.Data, &id); err != nil {
			return fmt.Errorf("unmarshal node id: %w", err)
		}
		delete(f.state.Nodes, id)

	case OpSetInstance:
		var inst Instance
		if err := json.Unmarshal(cmd.Data, &inst); err != nil {
			return fmt.Errorf("unmarshal instance: %w", err)
		}
		f.state.Instances[inst.ID] = inst

	case OpDeleteInstance:
		var id string
		if err := json.Unmarshal(cmd.Data, &id); err != nil {
			return fmt.Errorf("unmarshal instance id: %w", err)
		}
		delete(f.state.Instances, id)
	}

	return nil
}

func (f *FSM) GetApplication(name string) (Application, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	app, ok := f.state.Applications[name]
	return app, ok
}

func (f *FSM) GetAllApplications() []Application {
	f.mu.RLock()
	defer f.mu.RUnlock()
	apps := make([]Application, 0, len(f.state.Applications))
	for _, app := range f.state.Applications {
		apps = append(apps, app)
	}
	return apps
}

func (f *FSM) GetNode(id string) (Node, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.state.Nodes[id]
	return node, ok
}

func (f *FSM) GetAllNodes() []Node {
	f.mu.RLock()
	defer f.mu.RUnlock()
	nodes := make([]Node, 0, len(f.state.Nodes))
	for _, node := range f.state.Nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (f *FSM) GetInstancesForApp(appName string) []Instance {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var instances []Instance
	for _, inst := range f.state.Instances {
		if inst.AppName == appName {
			instances = append(instances, inst)
		}
	}
	return instances
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, err := json.Marshal(f.state)
	if err != nil {
		return nil, err
	}
	return &fsmSnapshot{data: data}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var state ClusterState
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}

	f.mu.Lock()
	f.state = state
	f.mu.Unlock()
	return nil
}

type fsmSnapshot struct {
	data []byte
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, bytes.NewReader(s.data)); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}
