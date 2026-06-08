package backend

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func alreadyLiked(postid string, userID string) bool {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	var count int

	err = db.QueryRow(`
	SELECT COUNT(*) 
	FROM post_like 
	WHERE post_id = ? AND user_id = ?;
	`, postid, userID).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func toggleLike(postid string, userID string) error {

	if alreadyLiked(postid, userID) {
		return deletelikepost(postid, userID)
	}

	return likepost(postid, userID)
}

func alreadyLikedComment(commentid string, userID string) bool {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	var count int

	err = db.QueryRow(`
	SELECT COUNT(*) 
	FROM comment_like 
	WHERE comments_id = ? AND user_id = ?;
	`, commentid, userID).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func toggleCommentLike(commentid string, userID string) error {
	if alreadyLikedComment(commentid, userID) {
		return deletelikecomment(commentid, userID)
	}

	return likecomment(commentid, userID)
}

func getPostLikeCount(postID string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM post_like WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func getCommentLikeCount(commentID string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comment_like WHERE comments_id = ?", commentID).Scan(&count)
	return count, err
}

func ToggleCommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.URL.Query().Get("id")

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = toggleCommentLike(commentID, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := getCommentLikeCount(commentID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"likes": count})
}

func ToggleLikeHandler(w http.ResponseWriter, r *http.Request) {

	postID := r.URL.Query().Get("id")

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = toggleLike(postID, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := getPostLikeCount(postID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"likes": count})
}
