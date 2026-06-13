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

func createPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	imagePath := ""
	theme := r.FormValue("theme")

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if !allowedTypes[ext] {
			http.Error(w, "Type de fichier non autorisé", http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Println(err)
			serverError(w)
			return
		}

		img, _, decodeErr := image.Decode(file)
		if decodeErr != nil {
			http.Error(w, "Impossible de décoder l'image", http.StatusBadRequest)
			return
		}

		resizedImg := imaging.Fit(img, maxImageSize, maxImageSize, imaging.Lanczos)

		imagePath = fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dstPath := filepath.Join(uploadDir, imagePath)
		dst, createErr := os.Create(dstPath)
		if createErr != nil {
			log.Println(err)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addPost(title, content, imagePath, theme, userID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showPosts(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	sessionUserID, err := getCurrentUserID(w, r)
	if err != nil || sessionUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	db, dbErr := sql.Open("sqlite", dbPath)
	if dbErr != nil {
		log.Println(dbErr)
		serverError(w)
		return
	}
	defer db.Close()

	var ownerID int
	if queryErr := db.QueryRow("SELECT user_id FROM posts WHERE id = ?;", id).Scan(&ownerID); queryErr != nil {
		http.Error(w, "Post introuvable", http.StatusNotFound)
		return
	}

	if ownerID != sessionUserID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	imagePath, imgErr := getImagePath(id)
	if imgErr == nil && imagePath != "" {
		os.Remove(filepath.Join(uploadDir, imagePath))
	}

	db.Exec("DELETE FROM comment_like WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id)
	db.Exec("DELETE FROM comment_dislike WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id)
	db.Exec("DELETE FROM comments WHERE post_id = ?;", id)
	db.Exec("DELETE FROM post_like WHERE post_id = ?;", id)
	db.Exec("DELETE FROM post_dislike WHERE post_id = ?;", id)
	db.Exec("DELETE FROM posts WHERE id = ?;", id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	userID, err := getCurrentUserID(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ownerID, err := getCommentOwnerID(id)
	if err != nil {
		http.Error(w, "Commentaire introuvable", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "Interdit", http.StatusForbidden)
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
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var ownerID int
	err = db.QueryRow("SELECT user_id FROM comments WHERE id = ?;", commentID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}

func deleteComment(commentID int) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM comments WHERE id = ?;", commentID)
	return err
}

func getImagePath(id int) (string, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var imagePath sql.NullString
	err = db.QueryRow("SELECT image_path FROM posts WHERE id = ?;", id).Scan(&imagePath)
	if err != nil {
		return "", err
	}
	if imagePath.Valid {
		return imagePath.String, nil
	}
	return "", nil
}
