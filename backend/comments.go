package backend

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func createCommentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		httpError(w, http.StatusUnauthorized)
		return
	}

	if err := addComment(postID, userID, content); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)

	if err != nil || postIDStr == "" {
		httpError(w, http.StatusBadRequest)
		return
	}

	currentUserID, _ := getCurrentUserID(w, r)

	comments, err := getComments(postID, currentUserID)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func editCommentHandler(w http.ResponseWriter, r *http.Request) {
	// On utilise ParseMultipartForm car le formulaire HTML envoie des données de ce type.
	// 10 << 20 correspond à une limite de 10MB.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		httpError(w, http.StatusUnauthorized)
		return
	}

	// --- Vérification de sécurité ---
	// On vérifie que l'utilisateur est bien le propriétaire du commentaire (ou un admin).
	ownerID, err := getCommentOwnerID(commentID)
	if err != nil {
		if err == sql.ErrNoRows {
			httpError(w, http.StatusNotFound) // Le commentaire n'existe pas.
		} else {
			serverError(w)
		}
		return
	}
	isAdmin, _ := getUserRole(userID)
	if ownerID != userID && !isAdmin {
		httpError(w, http.StatusForbidden) // L'utilisateur n'a pas les droits.
		return
	}

	if err := editComment(commentID, content, userID); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
}
