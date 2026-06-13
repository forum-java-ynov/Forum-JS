package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	var credentials LoginData
	if err := decodeRequest(r, &credentials); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	credentials.Username = strings.TrimSpace(credentials.Username)

	if credentials.Username == "" || credentials.Password == "" {
		httpError(w, http.StatusBadRequest)
		return
	}

	ok, err := loginUser(credentials.Username, credentials.Password)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	if !ok {
		httpError(w, http.StatusUnauthorized)
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
		httpError(w, http.StatusBadRequest)
		return
	}

	formData.FullName = strings.TrimSpace(formData.FullName)
	formData.Username = strings.TrimSpace(formData.Username)
	formData.Email = strings.TrimSpace(formData.Email)

	if formData.FullName == "" || formData.Username == "" || formData.Email == "" || formData.Password == "" || formData.ConfirmPassword == "" {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := insertUser(formData.FullName, formData.Username, formData.Email, formData.Password, formData.ConfirmPassword); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
