package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Node    NodeConfig    `yaml:"node"`
	Cluster ClusterConfig `yaml:"cluster"`
	API     APIConfig     `yaml:"api"`
	Proxy   ProxyConfig   `yaml:"proxy"`
	Log     LogConfig     `yaml:"log"`
}

type NodeConfig struct {
	Name          string `yaml:"name"`
	BindAddr      string `yaml:"bind_addr"`
	BindPort      int    `yaml:"bind_port"`
	AdvertiseAddr string `yaml:"advertise_addr"`
	DataDir       string `yaml:"data_dir"`
}

type ClusterConfig struct {
	Bootstrap bool     `yaml:"bootstrap"`
	JoinAddrs []string `yaml:"join_addrs"`
	JoinToken string   `yaml:"join_token"`
}

type APIConfig struct {
	GRPCPort int `yaml:"grpc_port"`
	HTTPPort int `yaml:"http_port"`
}

type ProxyConfig struct {
	HTTPPort  int `yaml:"http_port"`
	HTTPSPort int `yaml:"https_port"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := defaultAgentConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return cfg, nil
}

func defaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		Node: NodeConfig{
			BindAddr: "0.0.0.0",
			BindPort: 7946,
			DataDir:  "/var/lib/kudo",
		},
		Cluster: ClusterConfig{
			Bootstrap: false,
		},
		API: APIConfig{
			GRPCPort: 9090,
			HTTPPort: 8080,
		},
		Proxy: ProxyConfig{
			HTTPPort:  80,
			HTTPSPort: 443,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}
