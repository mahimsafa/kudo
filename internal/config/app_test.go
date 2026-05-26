package config

import (
	"testing"
)

func TestParseAppConfig(t *testing.T) {
	yaml := `
kind: Application
name: my-api
adapter: docker
replicas: 3
spec:
  image: myregistry/my-api:v1.0
  env:
    PORT: "8080"
  ports:
    - 8080
routing:
  domain: api.example.com
  path: /
  tls: auto
  algorithm: round-robin
  healthcheck:
    path: /health
    interval: 10s
    timeout: 3s
    unhealthy_threshold: 3
`
	apps, err := ParseAppConfigs([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].Name != "my-api" {
		t.Errorf("expected name 'my-api', got %q", apps[0].Name)
	}
	if apps[0].Replicas != 3 {
		t.Errorf("expected 3 replicas, got %d", apps[0].Replicas)
	}
	if apps[0].Spec.Image != "myregistry/my-api:v1.0" {
		t.Errorf("unexpected image: %s", apps[0].Spec.Image)
	}
	if apps[0].Routing.Domain != "api.example.com" {
		t.Errorf("unexpected domain: %s", apps[0].Routing.Domain)
	}
}

func TestParseMultipleApps(t *testing.T) {
	yaml := `
kind: Application
name: app1
adapter: docker
replicas: 2
spec:
  image: app1:latest
---
kind: Application
name: app2
adapter: nodejs
replicas: 1
spec:
  entrypoint: "npm start"
`
	apps, err := ParseAppConfigs([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
	if apps[0].Name != "app1" {
		t.Errorf("expected 'app1', got %q", apps[0].Name)
	}
	if apps[1].Name != "app2" {
		t.Errorf("expected 'app2', got %q", apps[1].Name)
	}
}
