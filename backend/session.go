package backend

import (
	"encoding/json"
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

func getCurrentUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := getSession(w, r)
	if err != nil {
		return "", err
	}
	userID, _ := session.Values["user_id"].(string)
	return userID, nil
}

func isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getCurrentUserID(w, r)
		if err != nil || userID == "" {
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
