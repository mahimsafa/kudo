package config

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Kind     string     `yaml:"kind"`
	Name     string     `yaml:"name"`
	Adapter  string     `yaml:"adapter"`
	Replicas int        `yaml:"replicas"`
	Spec     AppSpec    `yaml:"spec"`
	Routing  AppRouting `yaml:"routing"`
}

type AppSpec struct {
	Image      string            `yaml:"image,omitempty"`
	Entrypoint string            `yaml:"entrypoint,omitempty"`
	Directory  string            `yaml:"directory,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	Ports      []int             `yaml:"ports,omitempty"`
}

type AppRouting struct {
	Domain      string         `yaml:"domain,omitempty"`
	Path        string         `yaml:"path,omitempty"`
	TLS         string         `yaml:"tls,omitempty"`
	Algorithm   string         `yaml:"algorithm,omitempty"`
	HealthCheck HealthCheckCfg `yaml:"healthcheck,omitempty"`
}

type HealthCheckCfg struct {
	Path               string        `yaml:"path,omitempty"`
	Interval           time.Duration `yaml:"interval,omitempty"`
	Timeout            time.Duration `yaml:"timeout,omitempty"`
	UnhealthyThreshold int           `yaml:"unhealthy_threshold,omitempty"`
}

func ParseAppConfigs(data []byte) ([]AppConfig, error) {
	var apps []AppConfig

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var app AppConfig
		err := decoder.Decode(&app)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("parsing app config: %w", err)
		}
		if app.Kind != "Application" && app.Kind != "" {
			continue
		}
		if app.Name == "" {
			return nil, fmt.Errorf("application must have a name")
		}
		if app.Replicas == 0 {
			app.Replicas = 1
		}
		apps = append(apps, app)
	}

	return apps, nil
}
