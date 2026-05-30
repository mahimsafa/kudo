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
			Ports: config.PortList{{Port: 80}},
			Env:   map[string]string{"FOO": "bar"},
		},
		Routing: config.AppRouting{Domain: "demo.example.com", LocalAccess: true},
	})

	if app.Name != "nginx-demo" {
		t.Fatalf("name: got %q", app.Name)
	}
	if app.Spec["image"] != "nginx:alpine" {
		t.Fatalf("image spec: got %q", app.Spec["image"])
	}
	mappings := ParsePortMappingsFromSpec(app.Spec)
	if len(mappings) != 1 || mappings[0].Container != 80 {
		t.Fatalf("ports: got %v", mappings)
	}
	if !app.Routing.LocalAccess {
		t.Fatal("expected local_access")
	}
	env := ParseEnvFromSpec(app.Spec)
	if env["FOO"] != "bar" {
		t.Fatalf("env: got %v", env)
	}
}
