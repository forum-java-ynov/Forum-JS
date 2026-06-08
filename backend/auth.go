package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// -- Structs --

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterData struct {
	FullName        string `json:"full_name"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// -- OAuth config --

// oauthStateToken is used to prevent CSRF attacks.
const oauthStateToken = "csrf-state-token"

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes:       []string{"openid", "email", "profile"},
	Endpoint:     google.Endpoint,
}

// InitOAuth reloads the Google OAuth config from environment variables.
// must be called after .env is loaded
func InitOAuth() {
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// -- Auth handlers --

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials LoginData
	if err := decodeRequest(r, &credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	credentials.Username = strings.TrimSpace(credentials.Username)

	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	ok, err := loginUser(credentials.Username, credentials.Password)
	if err != nil {
		http.Error(w, "Login failed due to an internal error", http.StatusInternalServerError)
		return
	}

	if !ok {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	setSessionCookie(w, "user_email", credentials.Username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var formData RegisterData
	if err := decodeRequest(r, &formData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	formData.FullName = strings.TrimSpace(formData.FullName)
	formData.Username = strings.TrimSpace(formData.Username)
	formData.Email = strings.TrimSpace(formData.Email)

	if formData.FullName == "" || formData.Username == "" || formData.Email == "" ||
		formData.Password == "" || formData.ConfirmPassword == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if err := insertUser(formData.FullName, formData.Username, formData.Email, formData.Password, formData.ConfirmPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookiesToClear := []string{"user_email", "user_picture", "name"}
	for _, name := range cookiesToClear {
		http.SetCookie(w, &http.Cookie{
			Name:   name,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleMe returns the currently logged-in user's info as JSON.
func handleMe(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("user_email")
	if err != nil {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	picture := cookieValueOrEmpty(r, "user_picture")
	name := cookieValueOrEmpty(r, "name")
	if name == "" {
		name = emailCookie.Value
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"email":   emailCookie.Value,
		"picture": picture,
		"name":    name,
	})
}

// -- Google OAuth handlers --

// handleGoogleLogin redirects the user to Google's consent screen.
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL := googleOauthConfig.AuthCodeURL(oauthStateToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// handleGoogleCallback handles the redirect from Google after the user consents.
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateToken {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Failed to exchange authorization code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := fetchGoogleUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to fetch Google profile", http.StatusInternalServerError)
		return
	}

	if err := loginOrRegisterGoogleUser(userInfo); err != nil {
		http.Error(w, "Google login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	displayName := userInfo.Name
	if displayName == "" {
		displayName = userInfo.Email
	}

	setSessionCookie(w, "user_email", userInfo.Email)
	setSessionCookie(w, "user_picture", userInfo.Picture)
	setSessionCookie(w, "name", displayName)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// -- Helpers --

// fetchGoogleUserInfo calls the Google userinfo endpoint and returns the profile.
func fetchGoogleUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	client := googleOauthConfig.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// loginOrRegisterGoogleUser creates a new account if the Google user doesn't exist yet.
func loginOrRegisterGoogleUser(user *GoogleUserInfo) error {
	exists, err := userExistsByEmail(user.Email)
	if err != nil {
		return err
	}

	if !exists {
		return insertGoogleUser(user.Name, user.Email, user.ID)
	}

	return nil
}

// decodeRequest parses the request body into target, supporting both JSON and form data.
func decodeRequest(r *http.Request, target any) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return json.NewDecoder(r.Body).Decode(target)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	switch dst := target.(type) {
	case *LoginData:
		dst.Username = r.FormValue("username")
		dst.Password = r.FormValue("password")
	case *RegisterData:
		dst.FullName = r.FormValue("full_name")
		dst.Username = r.FormValue("username")
		dst.Email = r.FormValue("email")
		dst.Password = r.FormValue("password")
		dst.ConfirmPassword = r.FormValue("confirm_password")
	}

	return nil
}

// setSessionCookie sets an HttpOnly session cookie with sensible defaults.
func setSessionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// cookieValueOrEmpty returns the cookie value, or an empty string if it doesn't exist.
func cookieValueOrEmpty(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}
