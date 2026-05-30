package executor

import "context"

// PortMapping binds a container port to optional host and ingress ports.
type PortMapping struct {
	Container int
	Host      int
	Public    int
}

type DeployRequest struct {
	InstanceID string
	AppName    string
	Spec       map[string]string
	Env        map[string]string
	Ports      []PortMapping
}

type DeployResponse struct {
	Address string
	Status  string
}

type StopRequest struct {
	InstanceID string
}

type StatusResponse struct {
	InstanceID string
	Status     string
	Address    string
}

type HealthCheckResponse struct {
	Healthy bool
	Message string
}

type Adapter interface {
	Name() string
	Deploy(ctx context.Context, req DeployRequest) (*DeployResponse, error)
	Stop(ctx context.Context, req StopRequest) error
	Status(ctx context.Context, instanceID string) (*StatusResponse, error)
	HealthCheck(ctx context.Context, instanceID string) (*HealthCheckResponse, error)
	Logs(ctx context.Context, instanceID string) ([]string, error)
}
