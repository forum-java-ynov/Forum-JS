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

func ToggleLikeHandler(w http.ResponseWriter, r *http.Request) {

	postID := r.URL.Query().Get("id")

	err := toggleLike(postID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Like modifié"))
}
