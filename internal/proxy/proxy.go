package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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

func (p *Proxy) AddRoute(domain, path string, backends []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var urls []*url.URL
	for _, b := range backends {
		u, err := url.Parse(b)
		if err == nil {
			urls = append(urls, u)
		}
	}

	key := domain + path
	p.routes[key] = &route{backends: urls}
}

func (p *Proxy) RemoveRoute(domain, path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.routes, domain+path)
}

func (p *Proxy) UpdateBackends(domain, path string, backends []string) {
	p.AddRoute(domain, path, backends)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	key := r.Host + "/"
	rt, ok := p.routes[key]
	p.mu.RUnlock()

	if !ok {
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
	proxy.ServeHTTP(w, r)
}

func (p *Proxy) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, p)
}
