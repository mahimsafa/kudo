package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mahimsafa/kudo/internal/config"
	raftlayer "github.com/mahimsafa/kudo/internal/cluster/raft"
	"github.com/mahimsafa/kudo/internal/cluster/state"
)

func appConfigToState(app config.AppConfig) state.Application {
	spec := make(map[string]string)
	if app.Spec.Image != "" {
		spec["image"] = app.Spec.Image
	}
	if app.Spec.Entrypoint != "" {
		spec["entrypoint"] = app.Spec.Entrypoint
	}
	if app.Spec.Directory != "" {
		spec["directory"] = app.Spec.Directory
	}
	if len(app.Spec.Ports) > 0 {
		parts := make([]string, len(app.Spec.Ports))
		for i, p := range app.Spec.Ports {
			parts[i] = strconv.Itoa(p)
		}
		spec["ports"] = strings.Join(parts, ",")
	}
	if len(app.Spec.Env) > 0 {
		if b, err := json.Marshal(app.Spec.Env); err == nil {
			spec["env"] = string(b)
		}
	}

	routing := state.RoutingConfig{
		Domain:    app.Routing.Domain,
		Path:      app.Routing.Path,
		TLS:       app.Routing.TLS,
		Algorithm: app.Routing.Algorithm,
	}
	if app.Routing.HealthCheck.Path != "" {
		routing.HealthCheck = app.Routing.HealthCheck.Path
	}

	return state.Application{
		Name:     app.Name,
		Adapter:  app.Adapter,
		Replicas: app.Replicas,
		Spec:     spec,
		Routing:  routing,
	}
}

func applyYAMLToRaft(raft *raftlayer.RaftNode, yamlContent string, timeout time.Duration) (int, error) {
	apps, err := config.ParseAppConfigs([]byte(yamlContent))
	if err != nil {
		return 0, fmt.Errorf("parsing config: %w", err)
	}
	if len(apps) == 0 {
		return 0, fmt.Errorf("no applications found in config")
	}

	for _, appCfg := range apps {
		app := appConfigToState(appCfg)
		if existing, ok := raft.FSM().GetApplication(app.Name); ok {
			app.Version = existing.Version + 1
		} else {
			app.Version = 1
		}

		data, err := state.MarshalCommand(state.OpSetApplication, app)
		if err != nil {
			return 0, err
		}
		if err := raft.Apply(data, timeout); err != nil {
			return 0, fmt.Errorf("applying %q: %w", app.Name, err)
		}
	}

	return len(apps), nil
}

func ParsePortsFromSpec(spec map[string]string) []int {
	return parsePortsFromSpec(spec)
}

func ParseEnvFromSpec(spec map[string]string) map[string]string {
	return parseEnvFromSpec(spec)
}

func parsePortsFromSpec(spec map[string]string) []int {
	raw, ok := spec["ports"]
	if !ok || raw == "" {
		return nil
	}
	var ports []int
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		p, err := strconv.Atoi(part)
		if err == nil {
			ports = append(ports, p)
		}
	}
	return ports
}

func parseEnvFromSpec(spec map[string]string) map[string]string {
	raw, ok := spec["env"]
	if !ok || raw == "" {
		return nil
	}
	var env map[string]string
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return nil
	}
	return env
}
