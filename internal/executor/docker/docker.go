package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"

	"github.com/mahimsafa/kudo/internal/executor"
)

type DockerAdapter struct {
	client *client.Client
	logger *zap.Logger
}

func NewDockerAdapter(logger *zap.Logger) (*DockerAdapter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}

	return &DockerAdapter{client: cli, logger: logger}, nil
}

func (d *DockerAdapter) Name() string {
	return "docker"
}

func (d *DockerAdapter) Deploy(ctx context.Context, req executor.DeployRequest) (*executor.DeployResponse, error) {
	imageName := req.Spec["image"]
	if imageName == "" {
		return nil, fmt.Errorf("docker adapter requires 'image' in spec")
	}

	d.logger.Info("pulling image", zap.String("image", imageName))
	reader, err := d.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("pulling image: %w", err)
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, port := range req.Ports {
		p := nat.Port(fmt.Sprintf("%d/tcp", port))
		exposedPorts[p] = struct{}{}
		portBindings[p] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: ""}}
	}

	var envList []string
	for k, v := range req.Env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	instanceSuffix := req.InstanceID
	if len(instanceSuffix) > 8 {
		instanceSuffix = instanceSuffix[:8]
	}
	containerName := fmt.Sprintf("kudo-%s-%s", req.AppName, instanceSuffix)

	resp, err := d.client.ContainerCreate(ctx,
		&container.Config{
			Image:        imageName,
			Env:          envList,
			ExposedPorts: exposedPorts,
			Labels: map[string]string{
				"kudo.app":      req.AppName,
				"kudo.instance": req.InstanceID,
			},
		},
		&container.HostConfig{
			PortBindings: portBindings,
			RestartPolicy: container.RestartPolicy{
				Name: container.RestartPolicyUnlessStopped,
			},
		},
		nil, nil, containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("starting container: %w", err)
	}

	inspect, err := d.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	var address string
	for _, bindings := range inspect.NetworkSettings.Ports {
		if len(bindings) > 0 {
			address = fmt.Sprintf("127.0.0.1:%s", bindings[0].HostPort)
			break
		}
	}

	d.logger.Info("container started",
		zap.String("id", resp.ID[:12]),
		zap.String("name", containerName),
		zap.String("address", address),
	)

	return &executor.DeployResponse{
		Address: address,
		Status:  "running",
	}, nil
}

func (d *DockerAdapter) Stop(ctx context.Context, req executor.StopRequest) error {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing containers: %w", err)
	}

	for _, c := range containers {
		if c.Labels["kudo.instance"] == req.InstanceID {
			d.logger.Info("stopping container", zap.String("id", c.ID[:12]))
			if err := d.client.ContainerStop(ctx, c.ID, container.StopOptions{}); err != nil {
				return fmt.Errorf("stopping container: %w", err)
			}
			if err := d.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{}); err != nil {
				return fmt.Errorf("removing container: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("container for instance %s not found", req.InstanceID)
}

func (d *DockerAdapter) Status(ctx context.Context, instanceID string) (*executor.StatusResponse, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	for _, c := range containers {
		if c.Labels["kudo.instance"] == instanceID {
			status := "stopped"
			if strings.HasPrefix(c.State, "running") {
				status = "running"
			} else if c.State == "exited" {
				status = "failed"
			}

			return &executor.StatusResponse{
				InstanceID: instanceID,
				Status:     status,
			}, nil
		}
	}

	return &executor.StatusResponse{
		InstanceID: instanceID,
		Status:     "stopped",
	}, nil
}

func (d *DockerAdapter) HealthCheck(ctx context.Context, instanceID string) (*executor.HealthCheckResponse, error) {
	status, err := d.Status(ctx, instanceID)
	if err != nil {
		return &executor.HealthCheckResponse{Healthy: false, Message: err.Error()}, nil
	}

	return &executor.HealthCheckResponse{
		Healthy: status.Status == "running",
		Message: status.Status,
	}, nil
}

func (d *DockerAdapter) Logs(ctx context.Context, instanceID string) ([]string, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		if c.Labels["kudo.instance"] == instanceID {
			reader, err := d.client.ContainerLogs(ctx, c.ID, container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Tail:       "100",
			})
			if err != nil {
				return nil, err
			}
			defer reader.Close()
			data, _ := io.ReadAll(reader)
			return strings.Split(string(data), "\n"), nil
		}
	}

	return nil, fmt.Errorf("container not found for instance %s", instanceID)
}
