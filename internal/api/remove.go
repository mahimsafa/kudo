package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	pb "github.com/mahimsafa/kudo/internal/api/proto"
	"github.com/mahimsafa/kudo/internal/config"
	raftlayer "github.com/mahimsafa/kudo/internal/cluster/raft"
	"github.com/mahimsafa/kudo/internal/cluster/state"
	"github.com/mahimsafa/kudo/internal/executor"
)

type removeBlocker struct {
	AppName        string
	DependentApp   string
	Reason         string
	SharedResource string
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

func isMissingWorkload(err error) bool {
	return errors.Is(err, executor.ErrWorkloadNotFound)
}

func stopInstancesForApp(ctx context.Context, runtime *Runtime, app state.Application, instances []state.Instance) (warnings []string, err error) {
	for _, inst := range instances {
		if runtime == nil || runtime.StopInstance == nil || inst.NodeID != runtime.LocalNodeID {
			continue
		}
		if err := runtime.StopInstance(ctx, app.Adapter, inst.ID); err != nil {
			if isMissingWorkload(err) {
				warnings = append(warnings, fmt.Sprintf(
					"%q instance %s on node %q: container/workload not found (may have been removed manually)",
					app.Name, inst.ID, inst.NodeID,
				))
				continue
			}
			return warnings, fmt.Errorf("stopping instance %s for %q: %w", inst.ID, app.Name, err)
		}
	}
	return warnings, nil
}

func removeYAMLFromRaft(ctx context.Context, raft *raftlayer.RaftNode, runtime *Runtime, yamlContent string, forceMissing bool, timeout time.Duration) (*pb.RemoveResponse, error) {
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
		return &pb.RemoveResponse{
			Success:  false,
			Message:  "remove blocked by dependencies; no resources were changed",
			Blockers: blockersToProto(blockers),
		}, nil
	}

	var allWarnings []string
	for name, app := range toRemove {
		instances := fsm.GetInstancesForApp(name)
		warnings, err := stopInstancesForApp(ctx, runtime, app, instances)
		if err != nil {
			return &pb.RemoveResponse{Success: false, Message: err.Error()}, nil
		}
		allWarnings = append(allWarnings, warnings...)
	}

	if len(allWarnings) > 0 && !forceMissing {
		return &pb.RemoveResponse{
			Success:         false,
			ConfirmRequired: true,
			Warnings:        allWarnings,
			Message:         "some workloads are already gone; confirm to remove cluster state, routes, and remaining resources",
		}, nil
	}

	var removed []string
	for name, app := range toRemove {
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

	if runtime != nil && runtime.SyncProxy != nil {
		runtime.SyncProxy()
	}

	msg := fmt.Sprintf("removed %d application(s): %s", len(removed), strings.Join(removed, ", "))
	if len(notFound) > 0 {
		msg += fmt.Sprintf(" (not in cluster: %s)", strings.Join(notFound, ", "))
	}
	return &pb.RemoveResponse{
		Success:  true,
		Message:  msg,
		Warnings: allWarnings,
	}, nil
}
