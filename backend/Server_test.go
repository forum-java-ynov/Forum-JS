package backend

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../frontend/js"))))
	mux.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("../frontend"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/index.html")
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/register.html")
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/login.html")
	})
	mux.HandleFunc("/db/register", register)
	mux.HandleFunc("/db/login", login)

	return mux
}

func TestRouteIndex(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("/ : attendu 200, reçu %d", rr.Code)
	}
}

func TestRouteRegister(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/register", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("/register : attendu 200, reçu %d", rr.Code)
	}
}

func TestRouteLogin(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/login", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("/login : attendu 200, reçu %d", rr.Code)
	}
}

func TestRouteJs(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/js/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusInternalServerError {
		t.Errorf("/js/ : erreur serveur inattendue")
	}
}

func TestRouteDbRegisterMethodGet(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/db/register", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusInternalServerError {
		t.Errorf("/db/register : erreur serveur inattendue")
	}
}

func TestRouteDbLoginMethodGet(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/db/login", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusInternalServerError {
		t.Errorf("/db/login : erreur serveur inattendue")
	}
}
