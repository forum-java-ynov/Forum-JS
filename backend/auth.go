package backend

import (
	"encoding/json"
	"net/http"
	"strings"
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

	err := insertUser(
		data.Namecomplet,
		data.Username,
		data.Email,
		data.Password,
		data.VerifPassword,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
