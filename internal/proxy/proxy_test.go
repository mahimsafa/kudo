package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyRoutesbyDomain(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from backend"))
	}))
	defer backend.Close()

	p := NewProxy()
	p.AddRoute("test.example.com", "/", []string{backend.URL})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "test.example.com"
	rr := httptest.NewRecorder()

	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "hello from backend" {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestProxyLocalhostFallback(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	p := NewProxy()
	p.AddRoute("127.0.0.1", "/", []string{backend.URL})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "127.0.0.1:8088"
	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestProxyReturns502WhenNoBackends(t *testing.T) {
	p := NewProxy()
	p.AddRoute("test.example.com", "/", nil)

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "test.example.com"
	rr := httptest.NewRecorder()

	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}

func TestProxyReturns404ForUnknownDomain(t *testing.T) {
	p := NewProxy()

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "unknown.example.com"
	rr := httptest.NewRecorder()

	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestProxyRoundRobin(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	p := NewProxy()
	p.AddRoute("test.example.com", "/", []string{backend1.URL, backend2.URL})

	results := map[string]int{}
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Host = "test.example.com"
		rr := httptest.NewRecorder()
		p.ServeHTTP(rr, req)
		results[rr.Body.String()]++
	}

	if results["backend1"] != 2 || results["backend2"] != 2 {
		t.Errorf("expected round-robin distribution, got: %v", results)
	}
}
