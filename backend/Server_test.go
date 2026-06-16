package backend

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../frontend/js"))))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../frontend/css"))))
	mux.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("../frontend"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/register.html")
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/login.html")
	})
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	mux.HandleFunc("/db/register", rateLimiter(register, 5, time.Minute))
	mux.HandleFunc("/db/login", rateLimiter(login, 5, time.Minute))
	mux.HandleFunc("/db/create_post", isAuthenticated(createPostHandler))
	mux.HandleFunc("/db/posts", showPostsHandler)
	mux.HandleFunc("/db/delete_post", isAuthenticated(deletePostHandler))
	mux.HandleFunc("/db/edit-post", isAuthenticated(editPostHandler))
	mux.HandleFunc("/db/create_comment", isAuthenticated(createCommentHandler))
	mux.HandleFunc("/db/comments", showCommentsHandler)
	mux.HandleFunc("/db/edit_comment", isAuthenticated(editCommentHandler))
	mux.HandleFunc("/db/delete_comment", isAuthenticated(deleteCommentHandler))
	mux.HandleFunc("/db/toggle_like", isAuthenticated(ToggleLikeHandler))
	mux.HandleFunc("/db/toggle_dislike", isAuthenticated(ToggleDislikeHandler))
	mux.HandleFunc("/db/toggle_comment_like", isAuthenticated(ToggleCommentLikeHandler))
	mux.HandleFunc("/db/toggle_comment_dislike", isAuthenticated(ToggleCommentDislikeHandler))

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

func TestRouteNotFound(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest("GET", "/page-inexistante", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("/page-inexistante : attendu 404, reçu %d", rr.Code)
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

func TestRouteDbPostsGet(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	mux := newMux()
	req := httptest.NewRequest("GET", "/db/posts", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusInternalServerError {
		t.Errorf("/db/posts : erreur serveur inattendue")
	}
}

func TestRouteDbCommentsGet(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	mux := newMux()
	req := httptest.NewRequest("GET", "/db/comments", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusInternalServerError {
		t.Errorf("/db/comments : erreur serveur inattendue")
	}
}

func TestRouteProtectedRedirect(t *testing.T) {
	mux := newMux()
	routes := []string{
		"/db/create_post",
		"/db/delete_post",
		"/db/edit-post",
		"/db/create_comment",
		"/db/edit_comment",
		"/db/delete_comment",
		"/db/toggle_like",
		"/db/toggle_dislike",
		"/db/toggle_comment_like",
		"/db/toggle_comment_dislike",
	}

	for _, route := range routes {
		req := httptest.NewRequest("POST", route, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther && rr.Code != http.StatusUnauthorized {
			t.Errorf("%s sans auth : attendu 303 ou 401, reçu %d", route, rr.Code)
		}
	}
}
