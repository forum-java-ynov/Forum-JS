package backend

import (
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

	if err := validatePositiveID(postID); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")

	if err := validateCommentContent(content); err != nil {
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

	http.Redirect(w, r, "/?success=Commentaire+publié+avec+succès", http.StatusSeeOther)
}

func showCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)

	if err != nil || postIDStr == "" {
		httpError(w, http.StatusBadRequest)
		return
	}

	comments, err := getComments(postID)
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func editCommentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1 << 20); err != nil {
        if err := r.ParseForm(); err != nil {
            httpError(w, http.StatusBadRequest)
            return
        }
    }

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := validatePositiveID(commentID); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")

	if err := validateCommentContent(content); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		httpError(w, http.StatusUnauthorized)
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
	
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := validatePositiveID(id); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		httpError(w, http.StatusUnauthorized)
		return
	}

	ownerID, err := getCommentOwnerID(id)
	if err != nil {
		httpError(w, http.StatusNotFound)
		return
	}

	if ownerID != userID {
		httpError(w, http.StatusForbidden)
		return
	}

	if err := deleteComment(id); err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/?success=Commentaire+supprimé+avec+succès", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getCommentOwnerID(commentID int) (int, error) {
	var ownerID int
	err := DB.QueryRow("SELECT user_id FROM comments WHERE id = ?;", commentID).Scan(&ownerID)
	return ownerID, err
}

func deleteComment(commentID int) error {
	_, err := DB.Exec("DELETE FROM comments WHERE id = ?;", commentID)
	return err
}