package backend

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func alreadyLiked(postID int, userID int) bool {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM post_like WHERE post_id = ? AND user_id = ?;
	`, postID, userID).Scan(&count)
	return err == nil && count > 0
}

func alreadyDisliked(postID int, userID int) bool {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM post_dislike WHERE post_id = ? AND user_id = ?;
	`, postID, userID).Scan(&count)
	return err == nil && count > 0
}

func toggleLike(postID int, userID int) error {
	if alreadyLiked(postID, userID) {
		return deletePostLike(postID, userID)
	}
	if alreadyDisliked(postID, userID) {
		if err := deletePostDislike(postID, userID); err != nil {
			return err
		}
	}
	return insertPostLike(postID, userID)
}

func toggleDislike(postID int, userID int) error {
	if alreadyDisliked(postID, userID) {
		return deletePostDislike(postID, userID)
	}
	if alreadyLiked(postID, userID) {
		if err := deletePostLike(postID, userID); err != nil {
			return err
		}
	}
	return insertPostDislike(postID, userID)
}

func alreadyLikedComment(commentID int, userID int) bool {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM comment_like WHERE comments_id = ? AND user_id = ?;
	`, commentID, userID).Scan(&count)
	return err == nil && count > 0
}

func alreadyDislikedComment(commentID int, userID int) bool {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM comment_dislike WHERE comments_id = ? AND user_id = ?;
	`, commentID, userID).Scan(&count)
	return err == nil && count > 0
}

func toggleCommentLike(commentID int, userID int) error {
	if alreadyLikedComment(commentID, userID) {
		return deleteCommentLike(commentID, userID)
	}
	if alreadyDislikedComment(commentID, userID) {
		if err := deleteCommentDislike(commentID, userID); err != nil {
			return err
		}
	}
	return insertCommentLike(commentID, userID)
}

func toggleCommentDislike(commentID int, userID int) error {
	if alreadyDislikedComment(commentID, userID) {
		return deleteCommentDislike(commentID, userID)
	}
	if alreadyLikedComment(commentID, userID) {
		if err := deleteCommentLike(commentID, userID); err != nil {
			return err
		}
	}
	return insertCommentDislike(commentID, userID)
}

func getPostLikeCount(postID int) (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM post_like WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func getPostDislikeCount(postID int) (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM post_dislike WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func getCommentLikeCount(commentID int) (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM comment_like WHERE comments_id = ?", commentID).Scan(&count)
	return count, err
}

func getCommentDislikeCount(commentID int) (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM comment_dislike WHERE comments_id = ?", commentID).Scan(&count)
	return count, err
}

func ToggleCommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err = toggleCommentLike(commentID, userID); err != nil {
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
	commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err = toggleCommentDislike(commentID, userID); err != nil {
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
	postID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err = toggleLike(postID, userID); err != nil {
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
	postID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err = toggleDislike(postID, userID); err != nil {
		log.Println(err)
		serverError(w)
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
