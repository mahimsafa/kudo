package scheduler

import (
	"testing"

	"github.com/mahimsafa/kudo/internal/cluster/state"
)

func TestSchedulerSpread(t *testing.T) {
	s := NewScheduler()

	nodes := []state.Node{
		{ID: "node-1", Name: "node-1", Status: "healthy"},
		{ID: "node-2", Name: "node-2", Status: "healthy"},
		{ID: "node-3", Name: "node-3", Status: "healthy"},
	}

	existingInstances := []state.Instance{
		{ID: "i-1", AppName: "app1", NodeID: "node-1", Status: "running"},
	}

	picked, err := s.PickNode("app1", nodes, existingInstances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if picked == "node-1" {
		t.Error("expected scheduler to spread instances, but picked node-1 which already has an instance")
	}
}

func TestSchedulerSkipsUnhealthyNodes(t *testing.T) {
	s := NewScheduler()

	nodes := []state.Node{
		{ID: "node-1", Name: "node-1", Status: "unhealthy"},
		{ID: "node-2", Name: "node-2", Status: "healthy"},
	}

	picked, err := s.PickNode("app1", nodes, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if picked != "node-2" {
		t.Errorf("expected node-2 (healthy), got %s", picked)
	}
}

func TestSchedulerNoHealthyNodes(t *testing.T) {
	s := NewScheduler()

	nodes := []state.Node{
		{ID: "node-1", Name: "node-1", Status: "unhealthy"},
	}

	_, err := s.PickNode("app1", nodes, nil)
	if err == nil {
		t.Fatal("expected error when no healthy nodes available")
	}
}
