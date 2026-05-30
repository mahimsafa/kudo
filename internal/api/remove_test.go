package api

import (
	"testing"

	"github.com/hashicorp/raft"

	"github.com/mahimsafa/kudo/internal/cluster/state"
)

func applyApp(fsm *state.FSM, app state.Application) {
	cmd, _ := state.MarshalCommand(state.OpSetApplication, app)
	fsm.Apply(&raft.Log{Data: cmd})
}

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
		applyApp(fsm, app)
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

func TestRouteKey(t *testing.T) {
	key, ok := RouteKey(state.Application{
		Routing: state.RoutingConfig{Domain: "demo.example.com"},
	})
	if !ok || key != "demo.example.com/" {
		t.Fatalf("got %q ok=%v", key, ok)
	}
}
