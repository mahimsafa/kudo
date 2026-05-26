package scheduler

import (
	"fmt"
	"math"

	"github.com/mahimsafa/kudo/internal/cluster/state"
)

type Scheduler struct{}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) PickNode(appName string, nodes []state.Node, existingInstances []state.Instance) (string, error) {
	var healthyNodes []state.Node
	for _, n := range nodes {
		if n.Status == "healthy" {
			healthyNodes = append(healthyNodes, n)
		}
	}

	if len(healthyNodes) == 0 {
		return "", fmt.Errorf("no healthy nodes available")
	}

	nodeCount := make(map[string]int)
	for _, inst := range existingInstances {
		if inst.AppName == appName && (inst.Status == "running" || inst.Status == "starting") {
			nodeCount[inst.NodeID]++
		}
	}

	var bestNode string
	bestCount := math.MaxInt32

	for _, n := range healthyNodes {
		count := nodeCount[n.ID]
		if count < bestCount {
			bestCount = count
			bestNode = n.ID
		}
	}

	return bestNode, nil
}

func (s *Scheduler) PickNodes(appName string, count int, nodes []state.Node, existingInstances []state.Instance) ([]string, error) {
	var result []string
	instances := make([]state.Instance, len(existingInstances))
	copy(instances, existingInstances)

	for i := 0; i < count; i++ {
		nodeID, err := s.PickNode(appName, nodes, instances)
		if err != nil {
			return nil, err
		}
		result = append(result, nodeID)
		instances = append(instances, state.Instance{
			AppName: appName,
			NodeID:  nodeID,
			Status:  "starting",
		})
	}

	return result, nil
}
