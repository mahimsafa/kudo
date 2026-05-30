package api

import (
	"strings"

	"github.com/mahimsafa/kudo/internal/cluster/state"
)

// RouteKey returns a stable ingress key (domain+path) when routing.domain is set.
func RouteKey(app state.Application) (key string, ok bool) {
	domain := strings.TrimSpace(app.Routing.Domain)
	if domain == "" {
		return "", false
	}
	path := app.Routing.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return domain + path, true
}
