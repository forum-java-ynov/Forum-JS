package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const uploadDir = "uploads"
const maxImageSize = 800

var allowedTypes = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	theme := r.FormValue("theme")
	imagePath := ""

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		httpError(w, http.StatusUnauthorized)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if !allowedTypes[ext] {
			httpError(w, http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Println("Erreur de création de dossier:", err)
			serverError(w)
			return
		}

		img, _, decodeErr := image.Decode(file)
		if decodeErr != nil {
			httpError(w, http.StatusBadRequest)
			return
		}

		resizedImg := imaging.Fit(img, maxImageSize, maxImageSize, imaging.Lanczos)
		imagePath = fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dstPath := filepath.Join(uploadDir, imagePath)

		dst, createErr := os.Create(dstPath)
		if createErr != nil {
			log.Println("Erreur lors de la création du fichier image:", createErr)
			serverError(w)
			return
		}
		defer dst.Close()

		switch ext {
		case ".jpg", ".jpeg":
			jpeg.Encode(dst, resizedImg, &jpeg.Options{Quality: 85})
		case ".png":
			png.Encode(dst, resizedImg)
		}

		if ext == ".gif" || ext == ".webp" {
			file.Seek(0, 0)
			io.Copy(dst, file)
		}
	} else if err != http.ErrMissingFile {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := addPost(title, content, imagePath, theme, userID); err != nil {
		log.Println("Erreur lors de l'ajout du post:", err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := getPosts()
	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	sessionUserID, err := getCurrentUserID(w, r)
	if err != nil || sessionUserID == 0 {
		httpError(w, http.StatusUnauthorized)
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

	var ownerID int
	if queryErr := DB.QueryRow("SELECT user_id FROM posts WHERE id = ?;", id).Scan(&ownerID); queryErr != nil {
		httpError(w, http.StatusNotFound)
		return
	}

	if ownerID != sessionUserID {
		httpError(w, http.StatusUnauthorized)
		return
	}

	imagePath, imgErr := getImagePath(id)
	if imgErr == nil && imagePath != "" {
		os.Remove(filepath.Join(uploadDir, imagePath))
	}

	DB.Exec("DELETE FROM comment_like WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id)
	DB.Exec("DELETE FROM comment_dislike WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id)
	DB.Exec("DELETE FROM comments WHERE post_id = ?;", id)
	DB.Exec("DELETE FROM post_like WHERE post_id = ?;", id)
	DB.Exec("DELETE FROM post_dislike WHERE post_id = ?;", id)
	DB.Exec("DELETE FROM posts WHERE id = ?;", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

func getImagePath(id int) (string, error) {
	var imagePath sql.NullString
	err := DB.QueryRow("SELECT image_path FROM posts WHERE id = ?;", id).Scan(&imagePath)
	if err != nil {
		return "", err
	}
	if imagePath.Valid {
		return imagePath.String, nil
	}
	return "", nil
}
