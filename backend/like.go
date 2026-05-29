package backend

import (
	"database/sql"
	"net/http"
)

func alreadyLiked(postid string) bool {
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
	`, postid, 1).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func toggleLike(postid string) error {

	if alreadyLiked(postid) {
		return deletelikepost(postid)
	}

	return likepost(postid)
}

func alreadyLikedComment(commentid string) bool {
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
	`, commentid, 1).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func toggleCommentLike(commentid string) error {
	if alreadyLikedComment(commentid) {
		return deletelikecomment(commentid)
	}

	return likecomment(commentid)
}

func ToggleCommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.URL.Query().Get("id")

	err := toggleCommentLike(commentID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ToggleLikeHandler(w http.ResponseWriter, r *http.Request) {

	postID := r.URL.Query().Get("id")

	err := toggleLike(postID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
