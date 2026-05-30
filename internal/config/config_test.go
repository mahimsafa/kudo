package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgentConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "kudo.yaml")

	content := []byte(`
node:
  name: "test-node"
  bind_addr: "0.0.0.0"
  bind_port: 7946
  advertise_addr: "192.168.1.1"
  data_dir: "/var/lib/kudo"

cluster:
  bootstrap: true
  join_addrs: []

api:
  grpc_port: 9090
  http_port: 8080

proxy:
  http_port: 80
  https_port: 443

log:
  level: "info"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadAgentConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Node.Name != "test-node" {
		t.Errorf("expected node name 'test-node', got %q", cfg.Node.Name)
	}
	if cfg.Node.BindPort != 7946 {
		t.Errorf("expected bind port 7946, got %d", cfg.Node.BindPort)
	}
	if !cfg.Cluster.Bootstrap {
		t.Error("expected bootstrap to be true")
	}
}

func TestLoadAgentConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "kudo.yaml")

	content := []byte(`
node:
  name: "minimal"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadAgentConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Node.BindPort != 7946 {
		t.Errorf("expected default bind port 7946, got %d", cfg.Node.BindPort)
	}
	if cfg.API.GRPCPort != 9090 {
		t.Errorf("expected default grpc port 9090, got %d", cfg.API.GRPCPort)
	}
}

func TestLocalDevAgentConfig(t *testing.T) {
	cfg := LocalDevAgentConfig()
	if cfg.Node.BindAddr != "127.0.0.1" {
		t.Errorf("expected bind_addr 127.0.0.1, got %q", cfg.Node.BindAddr)
	}
	if cfg.Proxy.HTTPPort != 8088 {
		t.Errorf("expected proxy http_port 8088, got %d", cfg.Proxy.HTTPPort)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	wantData := filepath.Join(home, ".kudo", "data")
	if cfg.Node.DataDir != wantData {
		t.Errorf("expected data_dir %q, got %q", wantData, cfg.Node.DataDir)
	}
}
