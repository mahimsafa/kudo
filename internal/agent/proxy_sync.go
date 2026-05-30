package agent

import (
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/mahimsafa/kudo/internal/cluster/state"
	"github.com/mahimsafa/kudo/internal/proxy"
)

func (a *Agent) syncProxyRoutes() {
	if a.proxy == nil || a.raft == nil {
		return
	}
	if !a.raft.IsLeader() {
		return
	}

	fsm := a.raft.FSM()
	for _, app := range fsm.GetAllApplications() {
		domain := strings.TrimSpace(app.Routing.Domain)
		if domain == "" {
			continue
		}

		path := proxy.NormalizePath(app.Routing.Path)
		backends := instanceBackends(fsm, app.Name)
		if len(backends) == 0 {
			a.proxy.RemoveRoute(domain, path)
			if app.Routing.LocalAccess {
				a.proxy.RemoveRoute("localhost", path)
				a.proxy.RemoveRoute("127.0.0.1", path)
			}
			continue
		}

		a.proxy.UpdateBackends(domain, path, backends)
		if app.Routing.LocalAccess {
			a.proxy.UpdateBackends("localhost", path, backends)
			a.proxy.UpdateBackends("127.0.0.1", path, backends)
		}

		if app.Routing.IngressPort > 0 && app.Routing.IngressPort != a.config.Proxy.HTTPPort {
			a.logger.Debug("app ingress port differs from agent proxy listen port",
				zap.String("app", app.Name),
				zap.Int("ingress_port", app.Routing.IngressPort),
				zap.Int("proxy_listen_port", a.config.Proxy.HTTPPort),
			)
		}
	}
}

func instanceBackends(fsm *state.FSM, appName string) []string {
	var backends []string
	for _, inst := range fsm.GetInstancesForApp(appName) {
		if inst.Status != "running" || inst.Address == "" {
			continue
		}
		addr := inst.Address
		if !strings.Contains(addr, "://") {
			addr = "http://" + addr
		}
		backends = append(backends, addr)
	}
	return backends
}

func (a *Agent) logIngressHint(app state.Application) {
	port := app.Routing.IngressPort
	if port == 0 {
		port = a.config.Proxy.HTTPPort
	}
	domain := app.Routing.Domain
	if domain == "" {
		return
	}
	var access []string
	access = append(access, fmt.Sprintf("http://%s:%d%s (Host: %s)", "127.0.0.1", a.config.Proxy.HTTPPort, proxy.NormalizePath(app.Routing.Path), domain))
	if app.Routing.LocalAccess {
		access = append(access, fmt.Sprintf("http://127.0.0.1:%d%s", a.config.Proxy.HTTPPort, proxy.NormalizePath(app.Routing.Path)))
	}
	a.logger.Info("application reachable via L7 proxy",
		zap.String("app", app.Name),
		zap.Strings("urls", access),
		zap.Int("proxy_port", a.config.Proxy.HTTPPort),
	)
}
