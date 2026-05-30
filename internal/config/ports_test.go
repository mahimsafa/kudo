package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPortListUnmarshal(t *testing.T) {
	var spec struct {
		Ports PortList `yaml:"ports"`
	}
	data := `
ports:
  - 80
  - port: 8080
    public: 80
    host: 18080
`
	if err := yaml.Unmarshal([]byte(data), &spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Ports) != 2 {
		t.Fatalf("got %d ports", len(spec.Ports))
	}
	if spec.Ports[0].Port != 80 {
		t.Fatalf("first port: %+v", spec.Ports[0])
	}
	if spec.Ports[1].Port != 8080 || spec.Ports[1].Public != 80 || spec.Ports[1].Host != 18080 {
		t.Fatalf("second port: %+v", spec.Ports[1])
	}
}
