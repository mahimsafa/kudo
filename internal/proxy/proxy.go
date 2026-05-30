package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

type route struct {
	backends []*url.URL
	counter  uint64
}

type Proxy struct {
	mu     sync.RWMutex
	routes map[string]*route
}

func NewProxy() *Proxy {
	return &Proxy{
		routes: make(map[string]*route),
	}
}

// NormalizePath ensures paths are consistent route keys.
func NormalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func routeKey(host, path string) string {
	return host + NormalizePath(path)
}

func stripHostPort(host string) string {
	if h, _, ok := strings.Cut(host, ":"); ok {
		return h
	}
	return host
}

func (p *Proxy) AddRoute(domain, path string, backends []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.routes[routeKey(domain, path)] = p.buildRoute(backends)
}

func (p *Proxy) buildRoute(backends []string) *route {
	var urls []*url.URL
	for _, b := range backends {
		addr := b
		if !strings.Contains(addr, "://") {
			addr = "http://" + addr
		}
		u, err := url.Parse(addr)
		if err == nil {
			urls = append(urls, u)
		}
	}
	return &route{backends: urls}
}

func (p *Proxy) RemoveRoute(domain, path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.routes, routeKey(domain, path))
}

func (p *Proxy) UpdateBackends(domain, path string, backends []string) {
	p.AddRoute(domain, path, backends)
}

func (p *Proxy) lookupRoute(host, path string) *route {
	p.mu.RLock()
	defer p.mu.RUnlock()
	rt, ok := p.routes[routeKey(host, path)]
	if ok {
		return rt
	}
	for _, fallbackHost := range []string{"127.0.0.1", "localhost"} {
		if rt, ok = p.routes[routeKey(fallbackHost, path)]; ok {
			return rt
		}
	}
	return nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := stripHostPort(r.Host)
	path := NormalizePath(r.URL.Path)

	rt := p.lookupRoute(host, path)
	if rt == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if len(rt.backends) == 0 {
		http.Error(w, "no backends available", http.StatusBadGateway)
		return
	}

	idx := atomic.AddUint64(&rt.counter, 1) - 1
	target := rt.backends[idx%uint64(len(rt.backends))]

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
	proxy.ServeHTTP(w, r)
}

// ListenAndServe binds on all IPv4/IPv6 interfaces (0.0.0.0) so 127.0.0.1 works on macOS.
func (p *Proxy) ListenAndServe(addr string) error {
	if strings.HasPrefix(addr, ":") {
		addr = "0.0.0.0" + addr
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	return http.Serve(ln, p)
}
