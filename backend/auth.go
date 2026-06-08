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

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterData struct {
	Namecomplet   string `json:"full_name"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	VerifPassword string `json:"confirm_password"`
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Config OAuth Google

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes:       []string{"openid", "email", "profile"},
	Endpoint:     google.Endpoint,
}

// code de sécu pas sécu
const oauthStateToken = "etat-csrf-a-randomiser"

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	var data LoginData
	if err := decodeRequest(r, &data); err != nil {
		http.Error(w, "Requete invalide", http.StatusBadRequest)
		return
	}

	data.Username = strings.TrimSpace(data.Username)
	if data.Username == "" || data.Password == "" {
		http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
		return
	}

	result, err := loginUser(data.Username, data.Password)
	if err != nil {
		http.Error(w, "Erreur lors de la connexion", http.StatusInternalServerError)
		return
	}

	if result {
		session, err := getSession(w, r)
		if err != nil {
			http.Error(w, "Erreur de session", http.StatusInternalServerError)
			return
		}
		session.Values["user_id"] = data.Username
		session.Values["user_email"] = data.Username
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Erreur de session", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	var data RegisterData
	if err := decodeRequest(r, &data); err != nil {
		http.Error(w, "Requete invalide", http.StatusBadRequest)
		return
	}

	data.Namecomplet = strings.TrimSpace(data.Namecomplet)
	data.Username = strings.TrimSpace(data.Username)
	data.Email = strings.TrimSpace(data.Email)

	if data.Namecomplet == "" || data.Username == "" || data.Email == "" || data.Password == "" || data.VerifPassword == "" {
		http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
		return
	}

	if err := insertUser(data.Namecomplet, data.Username, data.Email, data.Password, data.VerifPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Handlers Google OAuth

func InitOauth() {
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// handleGoogleLogin redirect to google auth page
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateToken)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handleGoogleCallback gets google code and user infos
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("state") != oauthStateToken {
		http.Error(w, "State OAuth invalide", http.StatusBadRequest)
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Impossible d'échanger le code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := fetchGoogleUserInfo(token)
	if err != nil {
		http.Error(w, "Impossible de récupérer le profil Google", http.StatusInternalServerError)
		return
	}

	if err := loginOrRegisterGoogleUser(userInfo); err != nil {
		http.Error(w, "Erreur lors de la connexion Google: "+err.Error(), http.StatusInternalServerError)
		return
	}

	name := userInfo.Name
	if name == "" {
		name = userInfo.Email
	}

	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = userInfo.Email
	session.Values["user_email"] = userInfo.Email
	session.Values["user_name"] = name
	session.Values["user_picture"] = userInfo.Picture
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Impossible de sauvegarder la session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}

	id, _ := session.Values["user_id"].(string)
	if id == "" {
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
			name = id
		}
	}
	if email == "" {
		email = id
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"email":   email,
		"picture": picture,
		"name":    name,
	})
}

// fetchGoogleUserInfo calls google api to fetch profile
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

// loginOrRegisterGoogleUser insert user if doesn't exist
func loginOrRegisterGoogleUser(user *GoogleUserInfo) error {
	// Vérifie si l'utilisateur existe déjà via son email
	exists, err := userExistsByEmail(user.Email)
	if err != nil {
		return err
	}

	if !exists {
		return insertGoogleUser(user.Name, user.Email, user.ID)
	}

	return nil
}

func decodeRequest(r *http.Request, target any) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return json.NewDecoder(r.Body).Decode(target)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	switch data := target.(type) {
	case *LoginData:
		data.Username = r.FormValue("username")
		data.Password = r.FormValue("password")
	case *RegisterData:
		data.Namecomplet = r.FormValue("full_name")
		data.Username = r.FormValue("username")
		data.Email = r.FormValue("email")
		data.Password = r.FormValue("password")
		data.VerifPassword = r.FormValue("confirm_password")
	}

	return nil
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, "Erreur de session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Impossible de supprimer la session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
