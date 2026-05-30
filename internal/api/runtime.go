package api

import "context"

// Runtime performs workload and proxy cleanup on the local agent node.
type Runtime struct {
	LocalNodeID  string
	StopInstance func(ctx context.Context, adapter, instanceID string) error
	RemoveRoute  func(domain, path string)
}
