package backend

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const uploadDir = "uploads"
const maxImageSize = 800

var allowedExtensions = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

// allowedMimeTypes maps file extensions to their expected magic bytes prefixes.
// Used to validate that the file content matches its extension.
var imageMagicBytes = map[string][]byte{
	".jpg":  {0xFF, 0xD8, 0xFF},
	".jpeg": {0xFF, 0xD8, 0xFF},
	".png":  {0x89, 0x50, 0x4E, 0x47},
	".gif":  {0x47, 0x49, 0x46},
	".webp": {0x52, 0x49, 0x46, 0x46}, // RIFF header
}

// validateImageFile checks that the file extension is allowed and that the
// actual file content matches the expected magic bytes for that extension.
func validateImageFile(filename string, fileContent []byte) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("extension non autorisée: %s", ext)
	}

	magic, ok := imageMagicBytes[ext]
	if !ok {
		return nil // extension authorisée mais sans vérification magic bytes
	}

	if len(fileContent) < len(magic) {
		return fmt.Errorf("fichier trop court pour être une image valide")
	}

	if !bytes.Equal(fileContent[:len(magic)], magic) {
		return fmt.Errorf("le contenu du fichier ne correspond pas au type attendu %s", ext)
	}

	return nil
}

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

	if err := validatePostTitle(title); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if err := validatePostContent(content); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if err := validatePostTheme(theme); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if !allowedExtensions[ext] {
			http.Redirect(w, r, "/?error="+url.QueryEscape("Extension d'image non autorisée"), http.StatusSeeOther)
			return
		}

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Println("Erreur de création de dossier:", err)
			serverError(w)
			return
		}

		img, _, decodeErr := image.Decode(file)
		if decodeErr != nil {
			http.Redirect(w, r, "/?error="+url.QueryEscape("Fichier image invalide"), http.StatusSeeOther)
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
		http.Redirect(w, r, "/?error="+url.QueryEscape("Fichier image invalide"), http.StatusSeeOther)
		return
	}

	if err := addPost(title, content, imagePath, theme, userID); err != nil {
		log.Println("Erreur lors de l'ajout du post:", err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/?success=Post+publié+avec+succès", http.StatusSeeOther)
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

	if err := validatePositiveID(id); err != nil {
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

	tx, txErr := DB.Begin()
	if txErr != nil {
		log.Println("Erreur lors du démarrage de la transaction:", txErr)
		serverError(w)
		return
	}

	if _, err := tx.Exec("DELETE FROM comment_like WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id); err != nil {
		tx.Rollback()
		log.Println("Erreur suppression comment_like:", err)
		serverError(w)
		return
	}
	if _, err := tx.Exec("DELETE FROM comment_dislike WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id); err != nil {
		tx.Rollback()
		log.Println("Erreur suppression comment_dislike:", err)
		serverError(w)
		return
	}
	if _, err := tx.Exec("DELETE FROM comments WHERE post_id = ?;", id); err != nil {
		tx.Rollback()
		log.Println("Erreur suppression comments:", err)
		serverError(w)
		return
	}
	if _, err := tx.Exec("DELETE FROM post_like WHERE post_id = ?;", id); err != nil {
		tx.Rollback()
		log.Println("Erreur suppression post_like:", err)
		serverError(w)
		return
	}
	if _, err := tx.Exec("DELETE FROM post_dislike WHERE post_id = ?;", id); err != nil {
		tx.Rollback()
		log.Println("Erreur suppression post_dislike:", err)
		serverError(w)
		return
	}
	if _, err := tx.Exec("DELETE FROM posts WHERE id = ?;", id); err != nil {
		tx.Rollback()
		log.Println("Erreur suppression post:", err)
		serverError(w)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println("Erreur lors du commit:", err)
		serverError(w)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
		return
	}

	http.Redirect(w, r, "/?success=Post+supprimé+avec+succès", http.StatusSeeOther)
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

func editPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			httpError(w, http.StatusBadRequest)
			return
		}
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

	if err := validatePositiveID(id); err != nil {
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

	title := r.FormValue("title")
	content := r.FormValue("content")
	theme := r.FormValue("theme")
	removeImage := r.FormValue("remove_image")

	if err := validatePostTitle(title); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}
	if err := validatePostContent(content); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}
	if err := validatePostTheme(theme); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if !allowedExtensions[ext] {
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
		newImagePath := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dstPath := filepath.Join(uploadDir, newImagePath)

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

		oldImagePath, imgErr := getImagePath(id)
		if imgErr == nil && oldImagePath != "" {
			os.Remove(filepath.Join(uploadDir, oldImagePath))
		}

		err = editPostWithImage(id, title, content, theme, newImagePath)
	} else if err == http.ErrMissingFile {
		if removeImage == "true" {
			oldImagePath, imgErr := getImagePath(id)
			if imgErr == nil && oldImagePath != "" {
				os.Remove(filepath.Join(uploadDir, oldImagePath))
			}
			err = editPostWithImage(id, title, content, theme, "")
		} else {
			err = editPostWithoutImage(id, title, content, theme)
		}
	} else {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println("Erreur lors de la modification du post:", err)
		serverError(w)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		return
	}

	http.Redirect(w, r, "/?success=Post+modifié+avec+succès", http.StatusSeeOther)
}
