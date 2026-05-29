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
		http.SetCookie(w, &http.Cookie{
			Name:     "user_email",
			Value:    data.Username,
			Path:     "/",
			HttpOnly: true,
		})
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
	// Vérification de l'état CSRF
	if r.FormValue("state") != oauthStateToken {
		http.Error(w, "State OAuth invalide", http.StatusBadRequest)
		return
	}

	// trade code for access token
	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Impossible d'échanger le code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// getting userinfos from google
	userInfo, err := fetchGoogleUserInfo(token)
	if err != nil {
		http.Error(w, "Impossible de récupérer le profil Google", http.StatusInternalServerError)
		return
	}

	// Connexion ou inscription automatique de l'utilisateur
	if err := loginOrRegisterGoogleUser(userInfo); err != nil {
		http.Error(w, "Erreur lors de la connexion Google: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:     "user_email",
		Value:    userInfo.Email,
		Path:     "/",
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_picture",
		Value:    userInfo.Picture,
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	//check if user is logged in
	cookie, err := r.Cookie("user_email")
	if err != nil {
		http.Error(w, "Non connecté", http.StatusUnauthorized)
		return
	}
	pictureCookie, _ := r.Cookie("user_picture")
	picture := ""
	if pictureCookie != nil {
		picture = pictureCookie.Value
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"email":   cookie.Value,
		"picture": picture,
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
		// Crée le compte sans mot de passe (connexion Google uniquement)
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
	http.SetCookie(w, &http.Cookie{
		Name:   "user_email",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
