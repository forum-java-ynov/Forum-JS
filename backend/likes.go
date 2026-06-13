package backend

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

	if alreadyDisliked(postid, userID) {
		if err := deletedislikepost(postid, userID); err != nil {
			return err
		}
	}

	return likepost(postid, userID)
}

func alreadyDisliked(postid string, userID string) bool {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	var count int

	err = db.QueryRow(`
	SELECT COUNT(*) 
	FROM post_dislike 
	WHERE post_id = ? AND user_id = ?;
	`, postid, userID).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func toggleDislike(postid string, userID string) error {
	if alreadyDisliked(postid, userID) {
		return deletedislikepost(postid, userID)
	}

	if alreadyLiked(postid, userID) {
		if err := deletelikepost(postid, userID); err != nil {
			return err
		}
	}

	return dislikepost(postid, userID)
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

func alreadyDislikedComment(commentid string, userID string) bool {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	var count int

	err = db.QueryRow(`
	SELECT COUNT(*) 
	FROM comment_dislike 
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

	if alreadyDislikedComment(commentid, userID) {
		if err := deletedislikecomment(commentid, userID); err != nil {
			return err
		}
	}

	return likecomment(commentid, userID)
}

func toggleCommentDislike(commentid string, userID string) error {
	if alreadyDislikedComment(commentid, userID) {
		return deletedislikecomment(commentid, userID)
	}

	if alreadyLikedComment(commentid, userID) {
		if err := deletelikecomment(commentid, userID); err != nil {
			return err
		}
	}

	return dislikecomment(commentid, userID)
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

func getPostDislikeCount(postID string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM post_dislike WHERE post_id = ?", postID).Scan(&count)
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

func getCommentDislikeCount(commentID string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comment_dislike WHERE comments_id = ?", commentID).Scan(&count)
	return count, err
}

func ToggleCommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.URL.Query().Get("id")

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = toggleCommentLike(commentID, strconv.Itoa(userID))
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	newLikeCount, _ := getCommentLikeCount(commentID)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"likes": newLikeCount})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func ToggleCommentDislikeHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.URL.Query().Get("id")

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = toggleCommentDislike(commentID, strconv.Itoa(userID))
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	newDislikeCount, _ := getCommentDislikeCount(commentID)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"dislikes": newDislikeCount})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func ToggleLikeHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = toggleLike(postID, strconv.Itoa(userID))
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}
	newLikeCount, _ := getPostLikeCount(postID)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"likes": newLikeCount})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func ToggleDislikeHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = toggleDislike(postID, strconv.Itoa(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newDislikeCount, _ := getPostDislikeCount(postID)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"dislikes": newDislikeCount})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
