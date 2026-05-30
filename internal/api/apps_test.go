package api

import (
	"testing"

	"github.com/mahimsafa/kudo/internal/config"
)

func TestAppConfigToState(t *testing.T) {
	app := appConfigToState(config.AppConfig{
		Name:     "nginx-demo",
		Adapter:  "docker",
		Replicas: 2,
		Spec: config.AppSpec{
			Image: "nginx:alpine",
			Ports: []int{80},
			Env:   map[string]string{"FOO": "bar"},
		},
		Routing: config.AppRouting{Domain: "demo.example.com"},
	})

	if app.Name != "nginx-demo" {
		t.Fatalf("name: got %q", app.Name)
	}
	if app.Spec["image"] != "nginx:alpine" {
		t.Fatalf("image spec: got %q", app.Spec["image"])
	}
	ports := ParsePortsFromSpec(app.Spec)
	if len(ports) != 1 || ports[0] != 80 {
		t.Fatalf("ports: got %v", ports)
	}
	env := ParseEnvFromSpec(app.Spec)
	if env["FOO"] != "bar" {
		t.Fatalf("env: got %v", env)
	}
}
