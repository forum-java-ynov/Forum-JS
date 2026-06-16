package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2/github"
)

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

type User struct {
	ID       int
	FullName string
	Username string
	Email    string
}

// google
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes:       []string{"openid", "email", "profile"},
	Endpoint:     google.Endpoint,
}

//github

type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"` // Nom d'utilisateur GitHub
	Name      string `json:"name"`  // Nom complet
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

var githubOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
	Scopes:       []string{"read:user", "user:email"},
	Endpoint:     github.Endpoint,
}

func generateStateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func getUserByUsername(username string) (User, error) {
	var user User
	err := DB.QueryRow(
		"SELECT id, full_name, username, email FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.FullName, &user.Username, &user.Email)
	return user, err
}

func getUserByEmail(email string) (User, error) {
	var user User
	err := DB.QueryRow(
		"SELECT id, full_name, username, email FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.FullName, &user.Username, &user.Email)
	return user, err
}

func sessionDisplayName(values map[interface{}]interface{}) string {
	if name, _ := values["user_name"].(string); name != "" {
		return name
	}
	if email, _ := values["user_email"].(string); email != "" {
		return email
	}
	if id, _ := values["user_id"].(int); id != 0 {
		return fmt.Sprintf("%d", id)
	}
	return "Unknown"
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Méthode non autorisée"), http.StatusSeeOther)
		return
	}

	var credentials LoginData
	if err := decodeRequest(r, &credentials); err != nil {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Formulaire invalide"), http.StatusSeeOther)
		return
	}

	credentials.Username = strings.TrimSpace(credentials.Username)

	if credentials.Username == "" || credentials.Password == "" {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Tous les champs sont obligatoires"), http.StatusSeeOther)
		return
	}

	if err := validateUsername(credentials.Username); err != nil {
		http.Redirect(w, r, "/login?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	ok, err := loginUser(credentials.Username, credentials.Password)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	if !ok {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Identifiants incorrects"), http.StatusSeeOther)
		return
	}

	user, err := getUserByUsername(credentials.Username)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	session.Values["user_id"] = user.ID
	session.Values["user_email"] = user.Email
	session.Values["user_name"] = user.FullName
	if err := session.Save(r, w); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	var formData RegisterData
	if err := decodeRequest(r, &formData); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape("Formulaire invalide"), http.StatusSeeOther)
		return
	}

	formData.FullName = strings.TrimSpace(formData.FullName)
	formData.Username = strings.TrimSpace(formData.Username)
	formData.Email = strings.TrimSpace(formData.Email)

	if formData.FullName == "" || formData.Username == "" || formData.Email == "" || formData.Password == "" || formData.ConfirmPassword == "" {
		http.Redirect(w, r, "/register?error="+url.QueryEscape("Tous les champs sont obligatoires"), http.StatusSeeOther)
		return
	}

	if err := validateFullName(formData.FullName); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if err := validateUsername(formData.Username); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if err := validateEmail(formData.Email); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if err := validatePassword(formData.Password); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	if err := insertUser(formData.FullName, formData.Username, formData.Email, formData.Password, formData.ConfirmPassword); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login?success="+url.QueryEscape("Compte créé avec succès"), http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
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
		log.Println(err)
		serverError(w)
		return
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		httpError(w, http.StatusUnauthorized)
		return
	}

	email, _ := session.Values["user_email"].(string)
	picture, _ := session.Values["user_picture"].(string)
	name := sessionDisplayName(session.Values)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      fmt.Sprintf("%d", userID),
		"email":   email,
		"picture": picture,
		"name":    name,
	})
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateStateToken()
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	http.Redirect(w, r, googleOauthConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	expectedState, _ := session.Values["oauth_state"].(string)
	if expectedState == "" || r.FormValue("state") != expectedState {
		httpError(w, http.StatusBadRequest)
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	userInfo, err := fetchGoogleUserInfo(token)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	if err := loginOrRegisterGoogleUser(userInfo); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	user, err := getUserByEmail(userInfo.Email)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	displayName := userInfo.Name
	if displayName == "" {
		displayName = userInfo.Email
	}

	delete(session.Values, "oauth_state")
	session.Values["user_id"] = user.ID
	session.Values["user_email"] = user.Email
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
	_, err := DB.Exec(
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

// github
func handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	clientId := os.Getenv("GITHUB_CLIENT_ID")
	if clientId == "" {
		log.Println("❌ ERREUR CRITIQUE : GITHUB_CLIENT_ID est vide ! Vérifie tes variables d'environnement.")
		serverError(w)
		return
	}

	state, err := generateStateToken()
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	http.Redirect(w, r, githubOauthConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(w, r)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	expectedState, _ := session.Values["oauth_state"].(string)
	if expectedState == "" || r.FormValue("state") != expectedState {
		httpError(w, http.StatusBadRequest)
		return
	}

	token, err := githubOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	userInfo, err := fetchGitHubUserInfo(token)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	if err := loginOrRegisterGitHubUser(userInfo); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	user, err := getUserByEmail(userInfo.Email)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	delete(session.Values, "oauth_state")
	session.Values["user_id"] = user.ID
	session.Values["user_email"] = user.Email
	session.Values["user_name"] = userInfo.Name
	session.Values["user_picture"] = userInfo.AvatarURL

	if err := session.Save(r, w); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func fetchGitHubUserInfo(token *oauth2.Token) (*GitHubUserInfo, error) {
	client := githubOauthConfig.Client(context.Background(), token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []GitHubEmail
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary {
						userInfo.Email = e.Email
						break
					}
				}
			}
		}
	}

	// if no name, use username
	if userInfo.Name == "" {
		userInfo.Name = userInfo.Login
	}

	return &userInfo, nil
}

func loginOrRegisterGitHubUser(userInfo *GitHubUserInfo) error {
	exists, err := userExistsByEmail(userInfo.Email)
	if err != nil {
		return err
	}

	githubIDStr := fmt.Sprintf("%d", userInfo.ID)

	if !exists {
		return insertGitHubUser(userInfo.Name, userInfo.Email, githubIDStr)
	}

	return updateGitHubID(userInfo.Email, githubIDStr)
}
