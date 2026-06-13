package backend

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/rand"
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

func generateStateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes:       []string{"openid", "email", "profile"},
	Endpoint:     google.Endpoint,
}

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
		log.Println(err)
		serverError(w)
		return
	}

	if !ok {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	session.Values["user_id"] = credentials.Username
	session.Values["user_email"] = credentials.Username
	if sessionSaveErr := session.Save(r, w); sessionSaveErr != nil {
		log.Println(err)
		serverError(w)
		return
	}

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
	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}

	userID, _ := session.Values["user_id"].(string)
	if userID == "" {
		http.Error(w, "Non connecté", http.StatusUnauthorized)
		return
	}

	email, _ := session.Values["user_email"].(string)
	picture, _ := session.Values["user_picture"].(string)
	name, _ := session.Values["user_name"].(string)
	if name == "" {
		if email != "" {
			name = email
		} else {
			name = userID
		}
	}
	if email == "" {
		email = userID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      userID,
		"email":   email,
		"picture": picture,
		"name":    name,
	})
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL := googleOauthConfig.AuthCodeURL(oauthStateToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

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

	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = userInfo.ID
	session.Values["user_email"] = userInfo.Email
	session.Values["user_name"] = displayName
	session.Values["user_picture"] = userInfo.Picture
	if err := session.Save(r, w); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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

func updateGoogleID(email, googleID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"UPDATE users SET google_id = ? WHERE email = ? AND (google_id IS NULL OR google_id = '');",
		googleID, email,
	)
	return err
}

func loginOrRegisterGoogleUser(userInfo *GoogleUserInfo) error {
	exists, err := userExistsByEmail(userInfo.Email)
	if err != nil {
		return err
	}

	if !exists {
		return insertGoogleUser(userInfo.Name, userInfo.Email, userInfo.ID)
	}

	return updateGoogleID(userInfo.Email, userInfo.ID)
}

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

func setSessionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func cookieValueOrEmpty(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}
