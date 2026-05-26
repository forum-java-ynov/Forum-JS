package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const uploadDir = "uploads"

func createPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	imagePath := ""

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imagePath = filepath.Base(handler.Filename)
		dst, err := os.Create(filepath.Join(uploadDir, imagePath))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != http.ErrMissingFile {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addPost(title, content, imagePath)
	fmt.Fprint(w, "Post cree "+title+" "+content)
}

func showPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := getPosts()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, post := range posts {
		id := post["id"]

		comments, err := getComments(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(comments)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}