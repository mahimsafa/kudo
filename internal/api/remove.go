package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
	"github.com/mahimsafa/kudo/internal/config"
	raftlayer "github.com/mahimsafa/kudo/internal/cluster/raft"
	"github.com/mahimsafa/kudo/internal/cluster/state"
)

type removeBlocker struct {
	AppName         string
	DependentApp    string
	Reason          string
	SharedResource  string
}

func findRemoveBlockers(fsm *state.FSM, toRemove map[string]state.Application) []removeBlocker {
	var blockers []removeBlocker

	for name, app := range toRemove {
		key, hasRoute := RouteKey(app)
		if !hasRoute {
			continue
		}
		for _, other := range fsm.GetAllApplications() {
			if other.Name == name {
				continue
			}
			if _, removing := toRemove[other.Name]; removing {
				continue
			}
			otherKey, otherHasRoute := RouteKey(other)
			if otherHasRoute && otherKey == key {
				blockers = append(blockers, removeBlocker{
					AppName:        name,
					DependentApp:   other.Name,
					Reason:         "another application still uses the same ingress route",
					SharedResource: fmt.Sprintf("routing: %s", key),
				})
			}
		}
	}

	return blockers
}

func blockersToProto(blockers []removeBlocker) []*pb.DependencyInfo {
	out := make([]*pb.DependencyInfo, len(blockers))
	for i, b := range blockers {
		out[i] = &pb.DependencyInfo{
			AppName:        b.AppName,
			DependentApp:   b.DependentApp,
			Reason:         b.Reason,
			SharedResource: b.SharedResource,
		}
	}
	return out
}

func removeYAMLFromRaft(ctx context.Context, raft *raftlayer.RaftNode, runtime *Runtime, yamlContent string, timeout time.Duration) (*pb.RemoveResponse, error) {
	appCfgs, err := config.ParseAppConfigs([]byte(yamlContent))
	if err != nil {
		return &pb.RemoveResponse{Success: false, Message: fmt.Sprintf("parsing config: %v", err)}, nil
	}
	if len(appCfgs) == 0 {
		return &pb.RemoveResponse{Success: false, Message: "no applications found in config"}, nil
	}

	fsm := raft.FSM()
	toRemove := make(map[string]state.Application)
	var notFound []string

	for _, cfg := range appCfgs {
		app, exists := fsm.GetApplication(cfg.Name)
		if !exists {
			notFound = append(notFound, cfg.Name)
			continue
		}
		toRemove[cfg.Name] = app
	}

	if len(toRemove) == 0 {
		msg := "no matching applications in cluster"
		if len(notFound) > 0 {
			msg = fmt.Sprintf("%s (not found: %s)", msg, strings.Join(notFound, ", "))
		}
		return &pb.RemoveResponse{Success: true, Message: msg}, nil
	}

	blockers := findRemoveBlockers(fsm, toRemove)
	if len(blockers) > 0 {
		var lines []string
		for _, b := range blockers {
			lines = append(lines, fmt.Sprintf(
				"cannot remove %q: %s (blocked by %q, shared %s)",
				b.AppName, b.Reason, b.DependentApp, b.SharedResource,
			))
		}
		return &pb.RemoveResponse{
			Success:  false,
			Message:  "remove blocked by dependencies; no resources were changed",
			Blockers: blockersToProto(blockers),
		}, nil
	}

	var removed []string
	for name, app := range toRemove {
		instances := fsm.GetInstancesForApp(name)
		for _, inst := range instances {
			if runtime != nil && runtime.StopInstance != nil && inst.NodeID == runtime.LocalNodeID {
				if err := runtime.StopInstance(ctx, app.Adapter, inst.ID); err != nil {
					return &pb.RemoveResponse{
						Success: false,
						Message: fmt.Sprintf("stopping instance %s for %q: %v", inst.ID, name, err),
					}, nil
				}
			}
		}

		if runtime != nil && runtime.RemoveRoute != nil {
			if _, ok := RouteKey(app); ok {
				path := app.Routing.Path
				if path == "" {
					path = "/"
				}
				runtime.RemoveRoute(app.Routing.Domain, path)
			}
		}

		data, err := state.MarshalCommand(state.OpDeleteApplication, name)
		if err != nil {
			return &pb.RemoveResponse{Success: false, Message: err.Error()}, nil
		}
		if err := raft.Apply(data, timeout); err != nil {
			return &pb.RemoveResponse{Success: false, Message: fmt.Sprintf("removing %q: %v", name, err)}, nil
		}
		removed = append(removed, name)
	}

	msg := fmt.Sprintf("removed %d application(s): %s", len(removed), strings.Join(removed, ", "))
	if len(notFound) > 0 {
		msg += fmt.Sprintf(" (not in cluster: %s)", strings.Join(notFound, ", "))
	}
	return &pb.RemoveResponse{Success: true, Message: msg}, nil
}
