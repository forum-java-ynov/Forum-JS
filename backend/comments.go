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
	if err := r.ParseForm(); err != nil {
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
