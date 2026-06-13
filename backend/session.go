package backend

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func getSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "session-ID")
	if err != nil {
		return nil, err
	}
	return session, nil
}

var errNotAuthenticated = errors.New("not authenticated")

func getCurrentUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	session, err := getSession(w, r)
	if err != nil {
		return 0, err
	}

	id, ok := session.Values["user_id"].(int)
	if !ok || id == 0 {
		return 0, errNotAuthenticated
	}
	return id, nil
}

func isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getCurrentUserID(w, r)
		if err != nil || userID == 0 {
			if r.Header.Get("Accept") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
