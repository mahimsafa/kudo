package reconciler

import (
	"testing"

	"github.com/mahimsafa/kudo/internal/cluster/state"
)

func TestReconcileScaleUp(t *testing.T) {
	app := state.Application{Name: "test-app", Adapter: "docker", Replicas: 3}
	instances := []state.Instance{
		{ID: "i-1", AppName: "test-app", Status: "running", NodeID: "node-1"},
	}
	nodes := []state.Node{
		{ID: "node-1", Status: "healthy"},
		{ID: "node-2", Status: "healthy"},
	}

	actions := Reconcile(app, instances, nodes)

	scaleUpCount := 0
	for _, a := range actions {
		if a.Type == ActionDeploy {
			scaleUpCount++
		}
	}

	if scaleUpCount != 2 {
		t.Errorf("expected 2 deploy actions to scale from 1 to 3, got %d", scaleUpCount)
	}
}

func TestReconcileScaleDown(t *testing.T) {
	app := state.Application{Name: "test-app", Adapter: "docker", Replicas: 1}
	instances := []state.Instance{
		{ID: "i-1", AppName: "test-app", Status: "running", NodeID: "node-1"},
		{ID: "i-2", AppName: "test-app", Status: "running", NodeID: "node-2"},
		{ID: "i-3", AppName: "test-app", Status: "running", NodeID: "node-3"},
	}
	nodes := []state.Node{
		{ID: "node-1", Status: "healthy"},
		{ID: "node-2", Status: "healthy"},
		{ID: "node-3", Status: "healthy"},
	}

	actions := Reconcile(app, instances, nodes)

	stopCount := 0
	for _, a := range actions {
		if a.Type == ActionStop {
			stopCount++
		}
	}

	if stopCount != 2 {
		t.Errorf("expected 2 stop actions to scale from 3 to 1, got %d", stopCount)
	}
}

func TestReconcileNoChange(t *testing.T) {
	app := state.Application{Name: "test-app", Adapter: "docker", Replicas: 2}
	instances := []state.Instance{
		{ID: "i-1", AppName: "test-app", Status: "running", NodeID: "node-1"},
		{ID: "i-2", AppName: "test-app", Status: "running", NodeID: "node-2"},
	}
	nodes := []state.Node{
		{ID: "node-1", Status: "healthy"},
		{ID: "node-2", Status: "healthy"},
	}

	actions := Reconcile(app, instances, nodes)

	if len(actions) != 0 {
		t.Errorf("expected no actions when state matches desired, got %d", len(actions))
	}
}
