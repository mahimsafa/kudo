package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/raft"

	"github.com/mahimsafa/kudo/internal/cluster/state"
	"github.com/mahimsafa/kudo/internal/executor"
)

func TestFindRemoveBlockers_sharedRouting(t *testing.T) {
	fsm := state.NewFSM()

	web := state.Application{
		Name: "web", Adapter: "docker", Replicas: 1,
		Routing: state.RoutingConfig{Domain: "www.example.com", Path: "/"},
	}
	apiApp := state.Application{
		Name: "api", Adapter: "docker", Replicas: 1,
		Routing: state.RoutingConfig{Domain: "api.example.com", Path: "/"},
	}
	shared := state.Application{
		Name: "legacy", Adapter: "docker", Replicas: 1,
		Routing: state.RoutingConfig{Domain: "www.example.com", Path: "/"},
	}

	for _, app := range []state.Application{web, apiApp, shared} {
		cmd, _ := state.MarshalCommand(state.OpSetApplication, app)
		fsm.Apply(&raft.Log{Data: cmd})
	}

	toRemove := map[string]state.Application{"web": web}
	blockers := findRemoveBlockers(fsm, toRemove)
	if len(blockers) != 1 {
		t.Fatalf("expected 1 blocker, got %d", len(blockers))
	}
	if blockers[0].DependentApp != "legacy" {
		t.Fatalf("expected blocker legacy, got %q", blockers[0].DependentApp)
	}

	toRemoveBoth := map[string]state.Application{"web": web, "legacy": shared}
	if len(findRemoveBlockers(fsm, toRemoveBoth)) != 0 {
		t.Fatal("expected no blockers when removing all apps on shared route")
	}
}

func TestStopInstancesForApp_missingWorkloadWarning(t *testing.T) {
	runtime := &Runtime{
		LocalNodeID: "node-1",
		StopInstance: func(ctx context.Context, adapter, instanceID string) error {
			return fmt.Errorf("%w: container gone", executor.ErrWorkloadNotFound)
		},
	}
	app := state.Application{Name: "web", Adapter: "docker"}
	instances := []state.Instance{
		{ID: "inst-1", AppName: "web", NodeID: "node-1", Status: "running"},
	}

	warnings, err := stopInstancesForApp(context.Background(), runtime, app, instances)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %v", warnings)
	}
}

func TestIsMissingWorkload(t *testing.T) {
	err := fmt.Errorf("wrap: %w", executor.ErrWorkloadNotFound)
	if !isMissingWorkload(err) {
		t.Fatal("expected missing workload")
	}
	if isMissingWorkload(errors.New("other")) {
		t.Fatal("unexpected")
	}
}
