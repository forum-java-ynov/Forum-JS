package backend

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"

	ip := getIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("attendu '192.168.1.1', reçu '%s'", ip)
	}
}

func TestGetIP_NoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1"

	ip := getIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("attendu '192.168.1.1', reçu '%s'", ip)
	}
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	// Reset visitors pour ce test
	mu.Lock()
	visitors = make(map[string]*visitor)
	mu.Unlock()

	handler := rateLimiter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, 3, time.Minute)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("requête %d : attendu 200, reçu %d", i+1, rr.Code)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	mu.Lock()
	visitors = make(map[string]*visitor)
	mu.Unlock()

	handler := rateLimiter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, 3, time.Minute)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"

	// 3 premières requêtes OK
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		handler(rr, req)
	}

	// 4ème requête doit être bloquée
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("attendu 429, reçu %d", rr.Code)
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	mu.Lock()
	visitors = make(map[string]*visitor)
	mu.Unlock()

	handler := rateLimiter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, 1, time.Minute)

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "10.0.0.3:1111"

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.4:2222"

	rr1 := httptest.NewRecorder()
	handler(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("IP1 requête 1 : attendu 200, reçu %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	handler(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("IP2 requête 1 : attendu 200, reçu %d", rr2.Code)
	}

	// IP1 dépasse sa limite
	rr3 := httptest.NewRecorder()
	handler(rr3, req1)
	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 requête 2 : attendu 429, reçu %d", rr3.Code)
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	mu.Lock()
	visitors = make(map[string]*visitor)
	mu.Unlock()

	handler := rateLimiter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, 1, 100*time.Millisecond)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.5:1234"

	rr1 := httptest.NewRecorder()
	handler(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("requête 1 : attendu 200, reçu %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	handler(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("requête 2 : attendu 429, reçu %d", rr2.Code)
	}

	// Attendre que la fenêtre expire
	time.Sleep(150 * time.Millisecond)

	rr3 := httptest.NewRecorder()
	handler(rr3, req)
	if rr3.Code != http.StatusOK {
		t.Errorf("requête 3 (après reset) : attendu 200, reçu %d", rr3.Code)
	}
}
