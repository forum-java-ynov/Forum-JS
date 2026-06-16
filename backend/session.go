package backend

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"encoding/gob"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func init() {
	gob.Register(0)

	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   42400, // ~12h (in seconds)
	}
}

func setUserSession(userID int, sessionToken string) error {
	_, err := DB.Exec(`
        INSERT INTO user_sessions (user_id, session_token)
        VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET session_token = ?, created_at = CURRENT_TIMESTAMP
    `, userID, sessionToken, sessionToken)
	return err
}

func isSessionValid(userID int, sessionToken string) (bool, error) {
	var token string
	err := DB.QueryRow(
		"SELECT session_token FROM user_sessions WHERE user_id = ?", userID,
	).Scan(&token)
	if err != nil {
		return false, err
	}
	return token == sessionToken, nil
}

func deleteUserSession(userID int) error {
	_, err := DB.Exec("DELETE FROM user_sessions WHERE user_id = ?", userID)
	return err
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

	token, _ := session.Values["session_token"].(string)
	valid, err := isSessionValid(id, token)
	if err != nil || !valid {
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
