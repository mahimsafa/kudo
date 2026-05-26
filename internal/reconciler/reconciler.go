package reconciler

import (
	"context"
	"time"

	"go.uber.org/zap"

	raftlayer "github.com/mahimsafa/kudo/internal/cluster/raft"
	"github.com/mahimsafa/kudo/internal/cluster/state"
	"github.com/mahimsafa/kudo/internal/scheduler"
)

type ActionType string

const (
	ActionDeploy  ActionType = "deploy"
	ActionStop    ActionType = "stop"
	ActionRestart ActionType = "restart"
)

type Action struct {
	Type       ActionType
	AppName    string
	Adapter    string
	InstanceID string
	NodeID     string
}

func Reconcile(app state.Application, instances []state.Instance, nodes []state.Node) []Action {
	var actions []Action

	var activeInstances []state.Instance
	for _, inst := range instances {
		if inst.Status == "running" || inst.Status == "starting" {
			activeInstances = append(activeInstances, inst)
		}
	}

	activeCount := len(activeInstances)
	desired := app.Replicas

	if activeCount < desired {
		sched := scheduler.NewScheduler()
		needed := desired - activeCount
		nodeIDs, err := sched.PickNodes(app.Name, needed, nodes, instances)
		if err == nil {
			for _, nodeID := range nodeIDs {
				actions = append(actions, Action{
					Type:    ActionDeploy,
					AppName: app.Name,
					Adapter: app.Adapter,
					NodeID:  nodeID,
				})
			}
		}
	} else if activeCount > desired {
		excess := activeCount - desired
		for i := 0; i < excess && i < len(activeInstances); i++ {
			idx := len(activeInstances) - 1 - i
			actions = append(actions, Action{
				Type:       ActionStop,
				AppName:    app.Name,
				Adapter:    app.Adapter,
				InstanceID: activeInstances[idx].ID,
				NodeID:     activeInstances[idx].NodeID,
			})
		}
	}

	return actions
}

type ReconcileLoop struct {
	raft     *raftlayer.RaftNode
	logger   *zap.Logger
	interval time.Duration
	actionFn func(Action) error
}

func NewReconcileLoop(raft *raftlayer.RaftNode, logger *zap.Logger, interval time.Duration, actionFn func(Action) error) *ReconcileLoop {
	return &ReconcileLoop{
		raft:     raft,
		logger:   logger,
		interval: interval,
		actionFn: actionFn,
	}
}

func (r *ReconcileLoop) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !r.raft.IsLeader() {
				continue
			}
			r.reconcileAll()
		}
	}
}

func (r *ReconcileLoop) reconcileAll() {
	fsm := r.raft.FSM()
	apps := fsm.GetAllApplications()
	nodes := fsm.GetAllNodes()

	for _, app := range apps {
		instances := fsm.GetInstancesForApp(app.Name)
		actions := Reconcile(app, instances, nodes)

		for _, action := range actions {
			r.logger.Info("reconciler action",
				zap.String("type", string(action.Type)),
				zap.String("app", action.AppName),
				zap.String("node", action.NodeID),
			)
			if r.actionFn != nil {
				if err := r.actionFn(action); err != nil {
					r.logger.Error("action failed", zap.Error(err))
				}
			}
		}
	}
}
