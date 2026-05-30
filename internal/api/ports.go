package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/mahimsafa/kudo/internal/config"
	"github.com/mahimsafa/kudo/internal/executor"
)

func portMappingsToSpec(ports config.PortList) string {
	if len(ports) == 0 {
		return ""
	}
	mappings := make([]executor.PortMapping, len(ports))
	for i, p := range ports {
		mappings[i] = executor.PortMapping{
			Container: p.Port,
			Host:      p.Host,
			Public:    p.Public,
		}
	}
	b, err := json.Marshal(mappings)
	if err != nil {
		return ""
	}
	return string(b)
}

func ingressPortFromPorts(ports config.PortList, routing config.AppRouting) int {
	if routing.IngressPort > 0 {
		return routing.IngressPort
	}
	for _, p := range ports {
		if p.Public > 0 {
			return p.Public
		}
	}
	return 0
}

func ParsePortMappingsFromSpec(spec map[string]string) []executor.PortMapping {
	if raw, ok := spec["port_mappings"]; ok && raw != "" {
		var mappings []executor.PortMapping
		if err := json.Unmarshal([]byte(raw), &mappings); err == nil {
			return mappings
		}
	}
	// Legacy comma-separated container ports only.
	var out []executor.PortMapping
	for _, p := range parsePortsFromSpec(spec) {
		out = append(out, executor.PortMapping{Container: p})
	}
	return out
}

func primaryContainerPort(mappings []executor.PortMapping) int {
	if len(mappings) == 0 {
		return 0
	}
	return mappings[0].Container
}

// Legacy helper for tests.
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
