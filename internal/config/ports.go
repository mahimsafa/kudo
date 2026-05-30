package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// PortMapping describes a container port and optional host or ingress mapping.
type PortMapping struct {
	// Port is the port the application listens on inside the container.
	Port int `yaml:"port" json:"port"`
	// Host is an optional fixed port on the Docker host (0 = ephemeral).
	Host int `yaml:"host,omitempty" json:"host,omitempty"`
	// Public is the ingress port clients use on the Kudo L7 proxy (e.g. 80).
	Public int `yaml:"public,omitempty" json:"public,omitempty"`
}

// PortList accepts shorthand integers or mapping objects in YAML.
type PortList []PortMapping

func (p *PortList) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Kind != yaml.SequenceNode {
		return fmt.Errorf("ports must be a YAML list")
	}
	out := make(PortList, 0, len(value.Content))
	for i, node := range value.Content {
		switch node.Kind {
		case yaml.ScalarNode:
			var port int
			if err := node.Decode(&port); err != nil {
				return fmt.Errorf("ports[%d]: %w", i, err)
			}
			out = append(out, PortMapping{Port: port})
		case yaml.MappingNode:
			var m PortMapping
			if err := node.Decode(&m); err != nil {
				return fmt.Errorf("ports[%d]: %w", i, err)
			}
			if m.Port == 0 {
				return fmt.Errorf("ports[%d]: port is required", i)
			}
			out = append(out, m)
		default:
			return fmt.Errorf("ports[%d]: must be an integer or mapping", i)
		}
	}
	*p = out
	return nil
}
